package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"server/internal/constants"
	"strconv"
	"strings"
)

const (
	llmUrl     = "http://localhost:11434/api/generate"
	basePrompt = "You are an assistant that selects the most relevant Hadiths for the user's question based on the provided similarity scores.\n\n## Question:\n{InputText}\n\n## Candidates:\nEach hadith contains a similarity score (higher = more relevant), a book, page, and content.\n\nHadith {HadithID}\nScore: {Score}\nBook: {Book}\nPage: {Page}\nContent: {Content}\n\n=====\n\n(…repeat for each Hadith…)\n\n## Task:\n1. List the top 3 most relevant hadiths in order of similarity.\n2. For each, repeat this format:\n\nHadith {HadithID}\n{Content}\nSource: Hadith {HadithID}, Page {Page}, Book {Book}\nScore: {Score}\n\nIf unsure about relevance, say so explicitly, and make sure you always bring up the sources (perhaps analyse them yourself too to make sure they're actually relevant).\n"
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

func BuildPrompt(inputText string, inputVectors []float32, similarHadith []constants.HadithEmbeddingResponse) string {
	var promptBuilder strings.Builder
	promptBuilder.WriteString(basePrompt)
	promptBuilder.WriteString("Input: " + vectorToString(inputVectors) + "\n")
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

// vectorToString converts []float32 to a comma-separated string (to feed to the AI)
func vectorToString(vec []float32) string {
	builder := strings.Builder{}
	for i, val := range vec {
		if i > 0 {
			builder.WriteString(",")
		}
		builder.WriteString(fmt.Sprintf("%.6f", val))
	}
	return builder.String()
}
