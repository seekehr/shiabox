package main

import (
	"fmt"
	"server/llm"
)

func main() {
	fmt.Println("Sending prompt...")
	prompt := "hey thereeqweqwewq"
	resp, err := llm.SendPrompt(prompt)
	if err != nil {
		panic(err)
	}
	response, err := llm.ParseResponse(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("Model:", response)
}
