package main

import (
	"bufio"
	"fmt"
	"os"
	"server/internal/handlers/ai"
	"sync"
)

// Bismillah
func main() {
	reader := bufio.NewReader(os.Stdin)
	handler, err := handlers.NewHandler()
	if err != nil {
		panic(err)
	}
	for {
		fmt.Print("Enter your prompt: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			return
		}

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			response, err := handler.HandleRequest(input)
			if err != nil {
				panic(err)
			}
			fmt.Println("Model: " + response)
		}()
		wg.Wait()
	}
}
