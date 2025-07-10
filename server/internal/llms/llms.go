package llms

import (
	"fmt"
	"server/internal/constants"
	"server/internal/embedding"
	"strconv"
	"strings"
	"time"
)

const (
	chunkerPromptFile                   = "assets/books_parser_prompt.txt" // we dont want others to read this >.<
	chatPromptFile                      = "assets/prompt.txt"
	ParserModel        Model            = "gemma2-9b-it"
	ChatModel          Model            = "meta-llama/llama-4-scout-17b-16e-instruct"
	StreamedResponse   Stream           = true
	FullResponse       Stream           = false
	UserRole           Role             = "user"
	AssistantRole      Role             = "assistant"
	SystemRole         Role             = "system"
	StopFinishReason   FinishReasonType = "stop"
	LengthFinishReason FinishReasonType = "length"
	FilterFinishReason FinishReasonType = "content_filter"
)

type FinishReasonType string
type Stream bool
type Role string
type Model string

type LLM struct {
	ApiKey       string
	SystemPrompt string
}

type Handler struct {
	Groq *GroqLLM
}

// NewGlobalHandler - Create a new Handler to easily handle common use cases with the LLMs
func NewGlobalHandler(groq *GroqLLM) *Handler {
	return &Handler{Groq: groq}
}

// HandleChatRequest - Handle the chatting part of shiabox (the main part technically). Streamed response
func (handler *Handler) HandleChatRequest(prompt string) (<-chan *StreamedAIResponse, error) {
	start := time.Now()
	prompt = strings.TrimSpace(prompt)
	fmt.Println("\n\n====\nEmbedding prompt...")
	vectors, err := embedding.EmbedText(prompt, nil)
	if err != nil {
		fmt.Printf("Error embedding prompt: %v\n", err)
		return nil, err
	}

	fmt.Println("Prompt embedded of vec length " + strconv.Itoa(len(vectors)) + ". Now searching the prompt in the vector db...")
	// copies the pointer only; here to reduce handler.Groq bloat
	groq := handler.Groq
	foundVectors, err := groq.VectorDB.Search(vectors)
	if err != nil {
		return nil, err
	}

	fmt.Println(strconv.Itoa(len(foundVectors)) + " responses found.")
	fmt.Println("Building prompt...")
	parsedPrompt := BuildChatPrompt(prompt, foundVectors)
	fmt.Println("Prompt built and db searched in " + time.Since(start).String() + ".")
	fmt.Println("Sending prompt... (tokens: " + strconv.Itoa(len(parsedPrompt)) + " )")

	resp, err := groq.SendPrompt(parsedPrompt, true)
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

	return groq.Parser.ParseStreamedSSE(resp.Body), nil
}

func BuildChatPrompt(inputText string, similarHadith []constants.HadithEmbeddingResponse) string {
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

func BuildChunkerPrompt(inputText string) string {
	var promptBuilder strings.Builder
	promptBuilder.WriteString("\n" + inputText)
	return promptBuilder.String()
}
