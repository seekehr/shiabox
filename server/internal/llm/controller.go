package llm

import (
	"bytes"
	"encoding/json"
	"net/http"
)

const (
	llmUrl = "http://localhost:11434/api/generate"
)

// ollamaRequest Request format for the API
type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

func SendPrompt(prompt string) (*http.Response, error) {
	request := ollamaRequest{
		Model:  "mistral",
		Prompt: prompt,
	}
	parsedRequest, _ := json.Marshal(request)

	resp, err := http.Post(llmUrl, "application/json", bytes.NewBuffer(parsedRequest))
	return resp, err
}
