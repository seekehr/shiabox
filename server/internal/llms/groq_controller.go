package llms

// All about SENDING api requests to the LLM

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"server/internal/utils"
	"server/internal/vector"
)

const (
	llmUrl = "https://api.groq.com/openai/v1/chat/completions"
)

type GroqLLM struct {
	LLM
	Model    Model
	VectorDB *vector.Db
	Parser   *GroqParser
}

// AIMessage - Message API format for a message (`messages` for request, `delta` for response)
type AIMessage struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// promptRequest Request format for the API
type promptRequest struct {
	Messages []AIMessage `json:"messages"`
	Model    Model       `json:"model"`
	Stream   Stream      `json:"stream"`
}

func NewGroqHandler(model Model) (*GroqLLM, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GROQ_API_KEY env var not set")
	}

	vectorDb, err := vector.Connect()
	if err != nil {
		return nil, err
	}

	sysPrompt, err := utils.ReadTextFromFile(chatPromptFile)
	if err != nil {
		return nil, err
	}

	return &GroqLLM{
		LLM: LLM{
			ApiKey:       apiKey,
			SystemPrompt: sysPrompt,
		},
		VectorDB: vectorDb,
		Model:    model,
		Parser:   &GroqParser{},
	}, nil
}

func (groq *GroqLLM) SendPrompt(userPrompt string, streaming Stream) (*http.Response, error) {
	messages := []AIMessage{
		{
			Role:    SystemRole,
			Content: groq.SystemPrompt,
		},
		{
			Role:    UserRole,
			Content: userPrompt,
		},
	}

	request := promptRequest{
		Messages: messages,
		Model:    groq.Model,
		Stream:   streaming,
	}

	parsedRequest, _ := json.Marshal(request)
	return utils.MakeHeadersRequest(llmUrl, bytes.NewReader(parsedRequest), &http.Client{}, utils.Header{
		Key:   "Authorization",
		Value: "Bearer " + groq.ApiKey,
	}, utils.Header{
		Key:   "Content-Type",
		Value: "application/json",
	})
}
