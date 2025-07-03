package llm

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// OllamaRequest Request format for the API
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

const (
	llmUrl = "http://localhost:11434/api/generate"
)

func SendPrompt(prompt string) (*http.Response, error) {
	request := OllamaRequest{
		Model:  "phi",
		Prompt: prompt,
	}
	parsedRequest, _ := json.Marshal(request)

	resp, err := http.Post(llmUrl, "application/json", bytes.NewBuffer(parsedRequest))
	return resp, err
}
