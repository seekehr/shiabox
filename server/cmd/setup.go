package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/qdrant/go-client/qdrant"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"server/internal/constants"
	"server/internal/embedding"
	"server/internal/llms"
	"server/internal/utils"
	"server/internal/vector"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	FlagInitBooks        = 0
	FlagPostprocessBooks = 1
	FlagEmbedBooks       = 2
	FlagInitVectors      = 3
	FlagInitBoth         = 4
	FlagInitAll          = 5

	MaxVectors          = 50
	MaxVectorWorkers    = 10
	RatelimitSpeed      = 65    // to prevent ratelimit, in seconds. we use this value to sleep for the mins provided
	MaxRequestsPerMin   = 15    // max requests per minute FOR RATELIMIT.
	ChunkSizeCharacters = 50000 // in chars, not tokens. for the llm
	OverlapCharacters   = 2500  // characters we provide as context to LLM in case the quote is cut-off
)

type FinishedChunkJob struct {
	Index    int // for order
	Response string
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
			postprocessBooks()
		} else if flagged == FlagPostprocessBooks {
			postprocessBooks()
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

// chunkBooks - Use an LLM to convert .txt books into chunked JSON that can easily be fed into the vector db. Data is sorted for books, each book is done in order and each chunk is given to a goroutine to handle.
func chunkBooks(handler *llms.GeminiLLM) {
	timer := time.Now()
	files, err := os.ReadDir(constants.UnparsedBooksDir)
	if err != nil {
		panic(err)
	}

	processBook := func(book *os.File) {
		data, err := utils.ReadFileInChunks(book, ChunkSizeCharacters, OverlapCharacters)
		if err != nil {
			panic(err)
		}

		// if os.Open is ran with the full path of the book, the book.Name() will be fool path so we only get basename
		name := filepath.Base(book.Name())
		parseFile, err := os.OpenFile(constants.ParsedBooksDir+strings.Replace(name, ".txt", ".json", 1), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		defer parseFile.Close()

		fmt.Println("Sending requests to gemini for book " + name + ".")

		finishedJobs := make([]*FinishedChunkJob, 0)
		finishedJobsLock := &sync.Mutex{}
		index := 0           // for order
		requestsCounter := 0 // cuz we don't wanna reset index so cant use that; this is for rate-limits
		var wg sync.WaitGroup
		for chunk := range data {
			if requestsCounter >= MaxRequestsPerMin {
				time.Sleep(time.Second * RatelimitSpeed) // prevent rate limit
				fmt.Println("Slept for " + strconv.Itoa(RatelimitSpeed) + " seconds to prevent rate limit.")
				requestsCounter = 0
			}

			wg.Add(1)
			// send request to prompt
			go func(chunk string, index int) {
				defer wg.Done()
				fmt.Println("Spun up goroutine for chunk " + strconv.Itoa(index) + ".")
				resp, err := handler.SendPrompt(chunk)
				if err != nil {
					panic(err)
				}
				fmt.Println("Stop reason: " + string(resp.FinishReason))
				finishedJobsLock.Lock()
				finishedJobs = append(finishedJobs, &FinishedChunkJob{
					Index:    index,
					Response: resp.Content,
				})
				finishedJobsLock.Unlock()
				fmt.Println("Received response from Gemini for chunk " + strconv.Itoa(index) + ".")
			}(chunk, index) // fun fact: goroutines can access variables in the function scope that they are declared in.
			// we give them these params so they can copy them (to their reserved memory/scope) and it doesnt cause reference issue (i.e it uses index 2 because it points to index inside the loop instead of its index 0)
			index++
			requestsCounter++
		}

		wg.Wait()
		fmt.Println("Sorting chunks...")
		// sort from descending order
		sort.Slice(finishedJobs, func(i, j int) bool {
			return finishedJobs[i].Index < finishedJobs[j].Index
		})

		var content strings.Builder
		for _, job := range finishedJobs {
			content.WriteString(job.Response)
		}

		parseFile.Write([]byte(content.String()))
		fmt.Println("Written " + strconv.Itoa(len(content.String())) + " bytes to " + parseFile.Name() + ".")
		fmt.Println("Finished processing LLM response for " + name + ".")
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".txt") {
			continue
		}

		openedFile, err := os.Open(constants.UnparsedBooksDir + file.Name())
		if err != nil {
			panic(err)
		}
		os.Remove(constants.ParsedBooksDir + strings.Replace(file.Name(), ".txt", ".json", 1)) // remove the parsed file if it exists so we can overwrite it
		processBook(openedFile)
	}

	fmt.Println("Chunking done in: ", time.Since(timer).String())
}

// postprocessBooks - Remove junk from the books so they can have proper JSON syntax. This is needed due to LLM stupidity
func postprocessBooks() {
	timer := time.Now()
	files, err := os.ReadDir(constants.ParsedBooksDir)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			// we have to read file for contents so we can preprocess them, then overwrite the contents
			f, err := os.OpenFile(constants.ParsedBooksDir+file.Name(), os.O_RDWR, 0644)
			if err != nil {
				panic(err)
			}
			defer f.Close()

			// WHO CARES ABOUT MEMORY?!?!?!?! In all seriousness this will be bad if the server has lots of books and like <1G ram so maybe run on local machine before pushing to AWS or smth
			contents, err := io.ReadAll(f)
			if err != nil {
				panic(err)
			}

			// Clean the ][ because of LLM
			regex := regexp.MustCompile(`}\s*\]\s*\[\s*{`)
			cleaned := regex.ReplaceAll(contents, []byte("},{"))

			// NUKE FILE
			err = f.Truncate(0)
			if err != nil {
				panic(err)
			}

			// set file cursor to 0 so we can write from beginning
			_, err = f.Seek(0, 0)
			if err != nil {
				panic(err)
			}

			_, err = f.Write(cleaned)
			if err != nil {
				panic(err)
			}
		}()
	}
	wg.Wait()
	fmt.Println("Postprocessing books done in: ", time.Since(timer).String())
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
