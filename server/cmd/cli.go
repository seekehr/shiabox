package main

import (
	"bufio"
	"fmt"
	"os"
	"server/internal/embedding"
	"server/internal/llm"
	"server/internal/utils"
	"server/internal/vector"
	"strconv"
	"strings"
	"time"
)

// Bismillah
func main() {
	reader := bufio.NewReader(os.Stdin)
	llmPrompt, err := llm.ReadPrompt()
	if err != nil {
		panic(err)
	}
	vectorDb, err := vector.Connect()
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

		prompt := strings.TrimSpace(input)
		// we embed the prompt first so it can be searched
		fmt.Println("Embedding prompt...")
		vectors, err := embedding.EmbedText(prompt, nil)
		if err != nil {
			fmt.Printf("Error embedding prompt: %v\n", err)
			continue
		}

		fmt.Println("Prompt embedded of vec length " + strconv.Itoa(len(vectors)) + ". Now searching the prompt in the vector db...")
		foundVectors, err := vectorDb.Search(vectors)
		if err != nil {
			panic(err)
		}
		fmt.Println(strconv.Itoa(len(foundVectors)) + " responses found.")
		fmt.Println("Building prompt...")
		parsedPrompt := llm.BuildPrompt(llmPrompt, prompt, vectors, foundVectors)
		utils.SaveDataToDisk(parsedPrompt)
		fmt.Println("Sending prompt... (tokens: " + strconv.Itoa(len(parsedPrompt)) + " )")

		timer := time.Now()
		resp, err := llm.SendPrompt(parsedPrompt)
		if err != nil {
			fmt.Printf("Error sending prompt: %v\n", err)
			continue
		}

		response, err := llm.ParseResponse(resp.Body)
		if err != nil {
			fmt.Printf("Error parsing response: %v\n", err)
			continue
		}

		fmt.Println("Prompt sent in " + time.Since(timer).String() + ".")
		fmt.Println("Model:", response)
	}
}
