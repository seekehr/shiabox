package llm

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"server/internal/constants"
	"strconv"
	"strings"
)

const (
	llmUrl     = "http://localhost:11434/api/generate"
	promptFile = "assets/prompt.txt"
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

func BuildPrompt(llmPrompt string, inputText string, inputVectors []float32, similarHadith []constants.HadithEmbeddingResponse) string {
	var promptBuilder strings.Builder
	promptBuilder.WriteString(llmPrompt)
	promptBuilder.WriteString("InputText: " + inputText + "\n")
	promptBuilder.WriteString("<START>\n")
	for _, hadith := range similarHadith {
		promptBuilder.WriteString("Hadith: " + strconv.Itoa(hadith.Hadith) + "\n")
		promptBuilder.WriteString("Page: " + strconv.Itoa(hadith.Page) + "\n")
		promptBuilder.WriteString("Book: " + hadith.Book + "\n")
		promptBuilder.WriteString("Score: " + strconv.FormatFloat(float64(hadith.Score), 'f', -1, 32) + "\n")
		promptBuilder.WriteString("Content: " + hadith.Content + "\n")
		promptBuilder.WriteString("\n=====\n")
	}
	promptBuilder.WriteString("<END>\n")
	return promptBuilder.String()
}

func ReadPrompt() (string, error) {
	file, err := os.Open(promptFile)
	if err != nil {
		return "", err
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
