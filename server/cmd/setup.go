package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/qdrant/go-client/qdrant"
	"os"
	"server/internal/constants"
	"server/internal/embedding"
	"server/internal/llms"
	"server/internal/utils"
	"server/internal/vector"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	FlagInitBooks       = 0
	FlagEmbedBooks      = 1
	FlagInitVectors     = 2
	FlagInitBoth        = 3
	FlagInitAll         = 4
	MaxVectors          = 50
	MaxVectorWorkers    = 10
	RatelimitSpeed      = 4     // to prevent ratelimit, in seconds.
	ChunkSizeCharacters = 59000 // in chars, not tokens. for the llm
	OverlapCharacters   = 2500  // characters we provide as context to LLM in case the quote is cut-off
)

// ChunkJ*B

type ChunkJob struct {
	Book  string
	Lock  *sync.Mutex
	Chunk string
}

func main() {
	buf := bufio.NewScanner(os.Stdin)
	db, err := vector.Connect()
	if err != nil {
		panic(err)
	}

	gemini, err := llms.NewGeminiHandler(llms.ChunkerModel, context.Background(), llms.ChunkerPromptFile)
	if err != nil {
		panic(err)
	}

	fmt.Println("Would you like to parse books or initialise the vector db? Run each in order first if you're new. (0/1/2/3)")
	if buf.Scan() {
		flagged, err := strconv.Atoi(buf.Text())
		if err != nil {
			panic(err)
		}

		timeStart := time.Now()
		if flagged == FlagInitBooks {
			pdfToTxtBooks()
			chunkBooks(gemini)
		} else if flagged == FlagEmbedBooks {
			fmt.Println("Generating embeddings...")
			embedBooks()
		} else if flagged == FlagInitVectors {
			initVectors(db)
		} else if flagged == FlagInitBoth {
			fmt.Println("Generating embeddings...")
			embedBooks()
			initVectors(db)
		} else {
			panic("invalid flag: " + strconv.Itoa(flagged))
		}
		fmt.Println("Done in: ", time.Since(timeStart))
	}

}

// initBooks - Convert PDF books to .txt
func pdfToTxtBooks() {
	var parseWg sync.WaitGroup // because chunking relies on the .txt books so parsing them must be done first

	fmt.Println("MAKE SURE YOU HAVE PDFTOTEXT BY POPPLER'S UTILS INSTALLED.")
	fmt.Println("Converting pdf books to txt format...")
	timer := time.Now()

	// pdf books to convert to txt format
	pdfBooks, err := os.ReadDir(constants.PdfBooksDir)
	if err != nil {
		panic(err)
	}

	for _, entry := range pdfBooks {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".pdf") {
			continue
		}
		parseWg.Add(1)
		go func(fileName string) {
			defer parseWg.Done()
			parseError := utils.SavePDFAsTxt(constants.PdfBooksDir+name, constants.UnparsedBooksDir+strings.Replace(fileName, ".pdf", ".txt", 1))
			if parseError != nil {
				panic(parseError)
			}
		}(name)
	}

	parseWg.Wait()
	fmt.Println("Converting books to txt done in: ", time.Since(timer).String())
}

// chunkBooks - Use an LLM to convert .txt books into chunked JSON that can easily be fed into the vector db
func chunkBooks(handler *llms.GeminiLLM) {
	timer := time.Now()
	files, err := os.ReadDir(constants.UnparsedBooksDir)
	if err != nil {
		panic(err)
	}

	var jobsWg sync.WaitGroup

	jobsChan := make(chan *ChunkJob)
	go func() {
		for job := range jobsChan {
			jobsWg.Add(1)
			go func(job *ChunkJob) {
				defer jobsWg.Done()
				fmt.Println("Sending request to Gemini of length " + strconv.Itoa(len(job.Chunk)) + "...")
				resp, err := handler.SendPrompt(job.Chunk)
				if err != nil {
					panic(err)
				}
				fmt.Println("Stop reason: " + string(resp.FinishReason))
				fmt.Println("Received response from Gemini. Saving...")
				job.Lock.Lock()
				defer job.Lock.Unlock()
				err = os.WriteFile(constants.ParsedBooksDir+strings.Replace(job.Book, ".txt", ".json", 1), []byte(resp.Content), 0644)
				if err != nil {
					panic(err)
				}
				fmt.Println("Written response to file for job.")
			}(job)
			time.Sleep(time.Duration(RatelimitSpeed) * time.Second) // sleep to avoid ratelimit
		}
	}()

	// Go through each book and then forward the necessary chunk to our receiver
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".txt") {
			continue
		}

		openedFile, err := os.Open(constants.UnparsedBooksDir + file.Name())
		if err != nil {
			panic(err)
		}

		data, err := utils.ReadFileInChunks(openedFile, ChunkSizeCharacters, OverlapCharacters)
		if err != nil {
			panic(err)
		}
		lock := &sync.Mutex{}             // pass pointer to the same lock for each file
		for receivedChunk := range data { // send all jobs at once; the rate-limiting happens inside the jobs themselves
			jobsChan <- &ChunkJob{
				Book:  file.Name(),
				Chunk: receivedChunk,
				Lock:  lock,
			}
		}
	}

	jobsWg.Wait() // wait for all jobs to be finished
	fmt.Println("Chunking done in: ", time.Since(timer).String())
}

func embedBooks() {
	var wg sync.WaitGroup
	// read books in .json format awaiting to be embedded
	parsedBooks, _ := os.ReadDir(constants.ParsedBooksDir)

	for _, file := range parsedBooks {
		wg.Add(1)
		fmt.Println("Embedding one file: " + file.Name())

		go func(name string) {
			defer wg.Done()
			err := embedding.EmbedBook(constants.ParsedBooksDir+name, name)
			if err != nil {
				panic(err)
			}
		}(file.Name())
	}
	wg.Wait()
}

func initVectors(vectorDb *vector.Db) {
	var (
		doneWg sync.WaitGroup // a bit conflicted about keeping this for insert operations but better safe than sorry ig, as vector db might be closed or smth
		wg     sync.WaitGroup
	)

	embedDataChan := make(chan []constants.HadithEmbedding, MaxVectorWorkers*2)
	maxWorkersChan := make(chan int, MaxVectorWorkers)
	parsedBooks, _ := os.ReadDir(constants.EmbeddingsDir)

	/*
		this is smth interesting i had to learn. an unbuffered channel (i.e a channel without a `size` specification) blocks
		the code until a receiver is initialized. as such, we're supposed to init a receiver, as the channel will block until
		the data is received, but the data will not be received due to wg blocking until all sender goroutines are done,
		resulting in a deadlock.
		the reason we have this in a goroutine is to prevent blocking the main thread while listening for the channel.
		NOW, i did change the code to buffer the embedDataChan (`size` is initialized), but im still gonna keep this goroutine
		because the total size is not known. if the size is exceeded, then wg.Wait() will never be finished as the sender goroutines
		will block due to lack of space in the channel. it's also slower to send everything first and then process, than it is to
		process while data is being sent.
	*/

	doneWg.Add(1)
	go func() {
		defer doneWg.Done()
		var goroutineWg sync.WaitGroup
		for embed := range embedDataChan {
			maxWorkersChan <- 1
			goroutineWg.Add(1)
			go func(embedData []constants.HadithEmbedding) {
				defer func() {
					<-maxWorkersChan
					goroutineWg.Done()
				}()
				err := vectorDb.Add(embed)
				if err != nil {
					panic(err)
				}
			}(embed)
		}
		goroutineWg.Wait()
	}()

	for _, file := range parsedBooks {
		wg.Add(1)
		go func(fileName string) {
			defer wg.Done()
			// read all vectors from the embedded book
			embedData, err := embedding.ReadEmbeddedBook(constants.EmbeddingsDir + fileName)
			if err != nil {
				panic(err)
			}
			// read all the vectors in batches
			for i := 0; i < len(embedData); i += MaxVectors {
				end := i + MaxVectors
				if end > len(embedData) {
					end = len(embedData)
				}
				embedDataChan <- embedData[i:end]
			}
		}(file.Name())
	}

	wg.Wait()
	close(embedDataChan)
	doneWg.Wait() // can't use wg.Wait() as we want to close the channel (so the receiver goroutine stops listening), and THEN end the function (as all receiving goroutines are finished) but wg is used for sender goroutines to indicate SENDING is finished so can't use it
	// just for logging xd
	resp, err := vectorDb.Client.Count(context.Background(), &qdrant.CountPoints{
		CollectionName: vector.CollectionName,
		Exact:          proto.Bool(true), // ensures accurate count
	})

	if err != nil {
		panic(err)
	}

	fmt.Printf("Vector count: %d\n", resp.Result.Count)
}
