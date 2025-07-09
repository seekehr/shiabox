package llm

// All about SENDING api requests to the LLM

import (
	"bytes"
	"encoding/json"
	"net/http"
	"server/internal/constants"
	"server/internal/utils"
	"strconv"
	"strings"
)

const (
	llmUrl                 = "https://api.groq.com/openai/v1/chat/completions"
	parserPromptFile       = "assets/books_parser_prompt.txt"
	promptFile             = "assets/prompt.txt"
	ParserModel      Model = "llama-3.3-70b-versatile"
	ChatModel        Model = "meta-llama/llama-4-scout-17b-16e-instruct"
)

type stream bool
type Role string
type Model string

const (
	StreamedResponse stream = true
	FullResponse     stream = false
	UserRole         Role   = "user"
	AssistantRole    Role   = "assistant"
	SystemRole       Role   = "system"
)

// AIMessage - Message API format for a message (`messages` for request, `delta` for response)
type AIMessage struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// promptRequest Request format for the API
type promptRequest struct {
	Messages []AIMessage `json:"messages"`
	Model    Model       `json:"model"`
	Stream   stream      `json:"stream"`
}

func SendPrompt(sysPrompt string, userPrompt string, model Model, apiKey string, streaming stream) (*http.Response, error) {
	messages := []AIMessage{
		{
			Role:    UserRole,
			Content: prompt,
		},
	}

	request := promptRequest{
		Messages: messages,
		Model:    model,
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

func BuildPrompt(inputText string, similarHadith []constants.HadithEmbeddingResponse) string {
	var promptBuilder strings.Builder
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

func BuildParserPrompt(llmPrompt string, inputText string) string {
	var promptBuilder strings.Builder
	promptBuilder.WriteString(llmPrompt)
	promptBuilder.WriteString("\n" + inputText)
	return promptBuilder.String()
}

func ReadPrompt() (string, error) {
	return utils.ReadTextFromFile(promptFile)
}

func ReadParserPrompt() (string, error) {
	return utils.ReadTextFromFile(parserPromptFile)
}
