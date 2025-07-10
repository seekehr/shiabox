package handlers

import (
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
	"server/internal/embedding"
	"server/internal/llm"
	"server/internal/vector"
	"strconv"
	"strings"
	"time"
)

// AIHandler - Handles the entire User -> AI communication.
type AIHandler struct {
	llmApiKey string
	vectorDb  *vector.Db
	LlmPrompt string // only expose this as it is needed to prevent reading prompt from disk everytime
}

func NewHandler() (*AIHandler, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GROQ_API_KEY env var not set")
	}

	llmPrompt, err := llm.ReadPrompt()
	if err != nil {
		return nil, err
	}
	vectorDb, err := vector.Connect()
	if err != nil {
		return nil, err
	}

	return &AIHandler{
		llmApiKey: apiKey,
		vectorDb:  vectorDb,
		LlmPrompt: llmPrompt,
	}, nil
}

// HandleRequest - Handle the entire prompt -> AI process, and return the SSE stream of tokens
func (handler *AIHandler) HandleRequest(sysPrompt string, prompt string) (<-chan *llm.AIResponse, error) {
	start := time.Now()
	prompt = strings.TrimSpace(prompt)
	fmt.Println("\n\n====\nEmbedding prompt...")
	vectors, err := embedding.EmbedText(prompt, nil)
	if err != nil {
		fmt.Printf("Error embedding prompt: %v\n", err)
		return nil, err
	}

	fmt.Println("Prompt embedded of vec length " + strconv.Itoa(len(vectors)) + ". Now searching the prompt in the vector db...")
	foundVectors, err := handler.vectorDb.Search(vectors)
	if err != nil {
		return nil, err
	}

	fmt.Println(strconv.Itoa(len(foundVectors)) + " responses found.")
	fmt.Println("Building prompt...")
	parsedPrompt := llm.BuildChatPrompt(prompt, foundVectors)
	fmt.Println("Prompt built and db searched in " + time.Since(start).String() + ".")
	fmt.Println("Sending prompt... (tokens: " + strconv.Itoa(len(parsedPrompt)) + " )")

	resp, err := llm.SendPrompt(sysPrompt, parsedPrompt, llm.ChatModel, handler.llmApiKey, true)
	if resp != nil && resp.StatusCode == 429 {
		return nil, fmt.Errorf("ratelimit")
	}

	if err != nil {
		fmt.Printf("Error sending prompt: %v\n", err)
		return nil, err
	}

	if resp == nil {
		return nil, fmt.Errorf("nil response")
	}

	return llm.ParseStreamedSSE(resp.Body), nil
}

// HandleCompletePrompt - For non-streaming uses
func (handler *AIHandler) HandleCompletePrompt(sysPrompt string, prompt string, model llm.Model) (*llm.CompleteAIResponse, error) {
	resp, err := llm.SendPrompt(sysPrompt, prompt, model, handler.llmApiKey, false)
	if err != nil {
		return nil, err
	}
	fmt.Println("Status code " + strconv.Itoa(resp.StatusCode))
	defer resp.Body.Close()
	return llm.ParseResponse(resp.Body)
}

// GetSSEFlusher - Sets the headers to allow server-side events, and gives us the flusher to immediately push data
func GetSSEFlusher(c echo.Context) (http.Flusher, error) {
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().WriteHeader(200) // set our return code to 200

	/* the flusher is needed because http buffers our responses because an HTTP request for every small request would cause some
	performance issues */
	flusher, ok := c.Response().Writer.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("http req doesnt support sse")
	}

	return flusher, nil
}
