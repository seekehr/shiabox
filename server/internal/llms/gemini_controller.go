package llms

import (
	"context"
	_ "github.com/joho/godotenv"
	"google.golang.org/genai"
	"server/internal/utils"
)

type GeminiLLM struct {
	LLM
	Context context.Context
	Client  *genai.Client // BLESSING <3 LOVE U GEMINI (NOT GOOGLE THO I HATE GOOGLE)
	Parser  *GeminiParser
}

func NewGeminiHandler(model Model, ctx context.Context, sysPromptFile PromptFile) (*GeminiLLM, error) {
	// why does this need a ctx when we use one for requests :thonk:
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		return nil, err
	}

	sysPrompt, err := utils.ReadTextFromFile(string(sysPromptFile))
	if err != nil {
		return nil, err
	}

	return &GeminiLLM{
		LLM: LLM{
			ApiKey:       "", // client gets the api key from env's `GEMINI_API_KEY` automatically
			SystemPrompt: sysPrompt,
			Model:        model,
		},
		Context: ctx,
		Client:  client,
		Parser:  &GeminiParser{},
	}, nil
}

// SendPrompt - Send a prompt to the Gemini model. Does NOT support streaming.
func (gemini *GeminiLLM) SendPrompt(userPrompt string) (*GeminiCompleteAIResponse, error) {
	contents := make([]*genai.Content, 2)
	contents[0] = &genai.Content{
		Parts: []*genai.Part{ // pretty cool syntax honestly. we use * instead of & cuz we're defining the TYPE of the slice, and types obv dont use & as that gives pointer to a value
			{
				Text: gemini.SystemPrompt,
			},
		},
		Role: string(UserRole),
	}

	// we add user prompt after system prompt obv
	contents[1] = &genai.Content{
		Parts: []*genai.Part{
			{
				Text: userPrompt,
			},
		},
		Role: string(UserRole),
	}

	return gemini.Parser.ParseResponse(gemini.Client.Models.GenerateContent(gemini.Context, string(gemini.Model), contents, nil))
}
