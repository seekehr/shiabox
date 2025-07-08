package llm

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"server/internal/constants"
	"server/internal/utils"
	"strconv"
	"strings"
)

const (
	llmUrl     = "https://api.groq.com/openai/v1/chat/completions"
	promptFile = "assets/prompt.txt"
)

type stream bool

const (
	StreamedResponse stream = true
	FullResponse     stream = false
)

// message API format for a message
type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// promptRequest Request format for the API
type promptRequest struct {
	Messages []message `json:"messages"`
	Model    string    `json:"model"`
	Stream   stream    `json:"stream"`
}

func SendPrompt(prompt string, apiKey string, streaming stream) (*http.Response, error) {
	messages := make([]message, 1)
	messages = append(messages, message{
		Role:    "user",
		Content: prompt,
	})

	request := promptRequest{
		Messages: messages,
		Model:    "meta-llama/llama-4-scout-17b-16e-instruct",
		Stream:   streaming,
	}
	parsedRequest, _ := json.Marshal(request)
	return utils.MakeHeadersRequest(llmUrl, bytes.NewReader(parsedRequest), &http.Client{}, utils.Header{
		Key:   "Authorization",
		Value: "Bearer " + apiKey,
	}, utils.Header{
		Key:   "Content-Type",
		Value: "application/json",
	})
}

func BuildPrompt(llmPrompt string, inputText string, similarHadith []constants.HadithEmbeddingResponse) string {
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
