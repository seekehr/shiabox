package main

import (
	"bufio"
	"fmt"
	"os"
	"server/llm"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter your prompt: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			return
		}

		prompt := strings.TrimSpace(input)
		fmt.Println("Sending prompt...")

		resp, err := llm.SendPrompt(prompt)
		if err != nil {
			fmt.Printf("Error sending prompt: %v\n", err)
			continue
		}

		response, err := llm.ParseResponse(resp.Body)
		if err != nil {
			fmt.Printf("Error parsing response: %v\n", err)
			continue
		}

		fmt.Println("Model:", response)
	}
}
