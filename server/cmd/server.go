package main

import (
	"fmt"
	llm2 "server/internal/llm"
)

// Bismillah
func main() {
	fmt.Println("Sending prompt...")
	prompt := "hey thereeqweqwewq"
	resp, err := llm2.SendPrompt(prompt)
	if err != nil {
		panic(err)
	}
	response, err := llm2.ParseResponse(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("Model:", response)
}
