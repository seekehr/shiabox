package main

import (
	"bufio"
	"fmt"
	"os"
	"server/internal/constants"
	"server/internal/embedding"
	"server/internal/vector"
	"strconv"
	"sync"
	"time"
)

const (
	FlagParseBooks   = 0
	FlagInitVectors  = 1
	MaxVectors       = 50
	MaxVectorWorkers = 10
)

func main() {
	buf := bufio.NewScanner(os.Stdin)
	fmt.Println("Would you like to parse books or initialise the vector db? Run each in order first if you're new. (0/1)")
	if buf.Scan() {
		flagged, err := strconv.Atoi(buf.Text())
		if err != nil {
			panic(err)
		}

		fmt.Println("Generating embeddings...")
		timeStart := time.Now()

		if flagged == FlagParseBooks {
			parseBooks()
		} else if flagged == FlagInitVectors {
			db, err := vector.Connect()
			if err != nil {
				panic(err)
			}

			initVectors(db)
		} else {
			panic("invalid flag: " + strconv.Itoa(flagged))
		}
		fmt.Println("Done in: ", time.Since(timeStart))
	}

}

func parseBooks() {
	var wg sync.WaitGroup
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
	var wg sync.WaitGroup
	embedDataChan := make(chan []constants.HadithEmbedding)
	maxWorkersChan := make(chan int, MaxVectorWorkers)
	parsedBooks, _ := os.ReadDir(constants.EmbeddingsDir)

	for _, file := range parsedBooks {
		wg.Add(1)
		go func(fileName string) {
			defer wg.Done()
			embedData, err := embedding.ReadEmbeddedBook(constants.EmbeddingsDir + fileName)
			if err != nil {
				panic(err)
			}
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
	for embed := range embedDataChan {
		wg.Add(1)
		maxWorkersChan <- 1
		go func() {
			err := vectorDb.Add(embed)
			if err != nil {
				panic(err)
			}
			<-maxWorkersChan
		}()
	}
	wg.Wait()
}
