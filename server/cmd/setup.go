package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/qdrant/go-client/qdrant"
	"math"
	"os"
	"server/internal/constants"
	"server/internal/embedding"
	handlers "server/internal/handlers/ai"
	"server/internal/llm"
	"server/internal/utils"
	"server/internal/vector"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	FlagInitBooks        = 0
	FlagEmbedBooks       = 1
	FlagInitVectors      = 2
	FlagInitBoth         = 3
	FlagInitAll          = 4
	MaxParsers           = 30 // 30 rpm is the ratelimit for groq
	MaxVectors           = 50
	MaxVectorWorkers     = 10
	TokensSentPerRequest = 20000 // max completion token limit is 32k so this seems good for now ig
)

func main() {
	buf := bufio.NewScanner(os.Stdin)
	db, err := vector.Connect()
	if err != nil {
		panic(err)
	}
	parserPrompt, err := llm.ReadParserPrompt()
	if err != nil {
		panic(err)
	}

	handler, err := handlers.NewHandler()
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
			initBooks()
			chunkBooks(parserPrompt, handler)
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

func initBooks() {
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

// todo: chunkjob
func chunkBooks(parserPrompt string, handler *handlers.AIHandler) {
	// txt format books to chunk
	timer := time.Now()
	txtBooks, err := os.ReadDir(constants.UnparsedBooksDir)
	if err != nil {
		panic(err)
	}

	var (
		wg                 sync.WaitGroup
		activeWorkersMutex sync.Mutex
		firstRequest       *time.Time // used for tracking if 60 seconds have passed since the first request
	)

	// each file will have its own lock so we don't use another file's lock by accident
	fileLocks := make(map[string]*sync.Mutex)
	activeWorkers := 0

	for _, entry := range txtBooks {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".txt") {
			continue
		}

		fileContent, err := utils.ReadFileInChunks(constants.UnparsedBooksDir+entry.Name(), TokensSentPerRequest)
		if err != nil {
			panic(err)
		}

		jsonFile, err := os.OpenFile(constants.ParsedBooksDir+strings.Replace(entry.Name(), ".txt", ".json", 1), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}

		fileLocks[jsonFile.Name()] = &sync.Mutex{} // OUR file lock :DDD
		// we need to convert it to float for values like e.g 5.6 and then round up to account for left-over chunks
		numberOfRequestsNeeded := int(math.Ceil(float64(len(fileContent)) / float64(TokensSentPerRequest)))
		for i := 0; i < numberOfRequestsNeeded; i++ {
			if firstRequest != nil && time.Since(*firstRequest) < time.Second {
				if activeWorkers > MaxParsers {
					sleepDuration := time.Second - time.Since(*firstRequest)
					fmt.Println("Max Parsers reached, waiting for " + sleepDuration.String() + "...")
					time.Sleep(sleepDuration)
				}
			}

			if firstRequest == nil {
				currentTime := time.Now()
				firstRequest = &currentTime
				go func() {
					time.Sleep(time.Second)
					firstRequest = nil
					fmt.Println("Reset first request.")
				}()
			}

			activeWorkers += 1
			wg.Add(1)
			// communism
			go func(ourJsonFile *os.File, ourFileChannel <-chan string) {
				fmt.Println("Started worker number #" + strconv.Itoa(activeWorkers))
				defer func() {
					activeWorkersMutex.Lock()
					activeWorkers--
					activeWorkersMutex.Unlock()
					wg.Done()
				}()
				for promptTokens := range ourFileChannel {
					prompt := llm.BuildParserPrompt(promptTokens)
					resp, err := handler.HandleCompletePrompt(parserPrompt, prompt, llm.ParserModel)
					if err != nil {
						panic(err)
					}
					ourFileLock, exists := fileLocks[ourJsonFile.Name()]
					if !exists {
						panic("bruh wheres OUR file lock for " + ourJsonFile.Name())
					}
					ourFileLock.Lock()
					// o o o o o mg!!! WHAT IF INDEX OUT OF RANGE?!?! well, yolo (technically u live in hereafter too so yea)
					_, err = ourJsonFile.Write([]byte(resp.Choices[0].Message.Content + "\n"))
					if err != nil {
						panic(err)
					}
					ourFileLock.Unlock()
				}
			}(jsonFile, fileContent)
		}
		jsonFile.Close()
	}

	wg.Wait()
	fmt.Println("Chunking books done in: ", time.Since(timer).String())
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
