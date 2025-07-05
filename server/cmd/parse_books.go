package main

import (
	"fmt"
	"os"
	"server/internal/constants"
	"server/internal/embedding"
	"sync"
	"time"
)

func main() {
	fmt.Println("Generating embeddings...")
	var wg sync.WaitGroup
	timeStart := time.Now()
	parsedBooks, _ := os.ReadDir(constants.ParsedBooksDir)
	for _, file := range parsedBooks {
		wg.Add(1)
		fmt.Println("Embedding one file: " + file.Name())
		go func(name string) {
			err := embedding.EmbedBook(constants.ParsedBooksDir+name, name)
			if err != nil {
				panic(err)
			}
			wg.Done()
		}(file.Name())
	}
	wg.Wait()
	fmt.Println("Done in: ", time.Since(timeStart))
}
