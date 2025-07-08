package main

import (
	"bufio"
	"fmt"
	"os"
	"server/internal/handlers/ai"
	"time"
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

		dataStream, err := handler.HandleRequest(input)
		if err != nil {
			fmt.Printf("Error handling request: %v\n", err)
			return
		}

		timer := time.Now()
		fmt.Print("\nModel Response: ")
		for data := range dataStream {
			bestChoice := data.Choices[0]
			fmt.Print(bestChoice.Delta.Content)
		}
		fmt.Println()

		fmt.Println("Done in " + time.Since(timer).String() + ".")
		fmt.Print("Enter your prompt: ")
	}
}
