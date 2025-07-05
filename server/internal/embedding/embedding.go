package embedding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	url = "http://localhost:11434/api/embed"
)

type embeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type response struct {
	Embedding []float32 `json:"embeddings"`
}

func Embed(text string) ([]float32, error) {
	reqBody := embeddingRequest{
		Model: "mistral",
		Input: text,
	}

	// Marshal struct to JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var embeddings response
	json.NewDecoder(resp.Body).Decode(&embeddings)
	if embeddings.Embedding == nil {
		return nil, fmt.Errorf("empty embedding")
	}
	return embeddings.Embedding, err
}
