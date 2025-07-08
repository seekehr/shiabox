package handlers

import (
	"fmt"
	"server/internal/embedding"
	"server/internal/llm"
	"server/internal/vector"
	"strconv"
	"strings"
	"time"
)

// Handler Handles the entire User -> AI communication
type Handler struct {
	vectorDb  *vector.Db
	llmPrompt string
}

func NewHandler() (*Handler, error) {
	llmPrompt, err := llm.ReadPrompt()
	if err != nil {
		return nil, err
	}
	vectorDb, err := vector.Connect()
	if err != nil {
		return nil, err
	}

	return &Handler{
		vectorDb:  vectorDb,
		llmPrompt: llmPrompt,
	}, nil
}

func (handler *Handler) HandleRequest(prompt string) (chan string, error) {
	start := time.Now()
	prompt = strings.TrimSpace(prompt)
	fmt.Println("\n\n====\nEmbedding prompt...")
	vectors, err := embedding.EmbedText(prompt, nil)
	if err != nil {
		fmt.Printf("Error embedding prompt: %v\n", err)
		return nil, err
	}

	fmt.Println("Prompt embedded of vec length " + strconv.Itoa(len(vectors)) + ". Now searching the prompt in the vector db...")
	foundVectors, err := handler.vectorDb.Search(vectors)
	if err != nil {
		return nil, err
	}

	fmt.Println(strconv.Itoa(len(foundVectors)) + " responses found.")
	fmt.Println("Building prompt...")
	parsedPrompt := llm.BuildPrompt(handler.llmPrompt, prompt, foundVectors)
	fmt.Println("Prompt built and db searched in " + time.Since(start).String() + ".")
	fmt.Println("Sending prompt... (tokens: " + strconv.Itoa(len(parsedPrompt)) + " )")

	resp, err := llm.SendPrompt(parsedPrompt, true)
	if err != nil {
		fmt.Printf("Error sending prompt: %v\n", err)
		return nil, err
	}

	return llm.ParseStreamedResponse(resp.Body), nil
}
