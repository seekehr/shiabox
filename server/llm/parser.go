package llm

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

const (
	bufferSize = 2 * 1024 * 1024 // 2 mib
)

type ResponseChunk struct {
	Response string `json:"response"`
}

func ParseResponse(body *io.ReadCloser) (string, error) {
	defer (*body).Close()
	var fullResponse string
	var chunk ResponseChunk

	scanner := bufio.NewScanner(*body)
	scanner.Buffer(make([]byte, 0, 64*1024), bufferSize)
	for scanner.Scan() {
		line := scanner.Text()
		if err := json.Unmarshal([]byte(line), &chunk); err == nil {
			fullResponse += chunk.Response
		} else {
			return "", err
		}
	}
	if scanner.Err() != nil {
		return "", scanner.Err()
	}
	if strings.TrimSpace(fullResponse) == "" || fullResponse == "" {
		return "", fmt.Errorf("empty response")
	}

	return fullResponse, nil
}
