package llm

// All about SENDING api requests to the LLM

import (
	"bytes"
	"encoding/json"
	"net/http"
	"server/internal/utils"
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
			Role:    SystemRole,
			Content: sysPrompt,
		},
		{
			Role:    UserRole,
			Content: userPrompt,
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
