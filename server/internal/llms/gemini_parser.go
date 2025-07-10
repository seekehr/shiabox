package llms

import (
	"fmt"
	"google.golang.org/genai"
)

// Could it be merged with controller? Yes, but consistency is key (ig......)

type GeminiParser struct{} // empty struct for method usage uwu

// GeminiCompleteAIResponse - Long name ;3. Content from gemini request is properly structured for convenience
type GeminiCompleteAIResponse struct {
	Content      string
	FinishReason genai.FinishReason // Different from the custom defined groq typ
}

// ParseResponse - Parse an unstreamed response from the API request made to Gemini's LLM.
func (GeminiParser) ParseResponse(resp *genai.GenerateContentResponse, err error) (*GeminiCompleteAIResponse, error) {
	if err != nil {
		return nil, err // judging anyone for bad code is not good
	}

	candidate := resp.Candidates[0]
	if len(resp.Candidates) < 1 {
		return nil, fmt.Errorf("no candidates found")
	}

	if len(candidate.Content.Parts) < 1 {
		return nil, fmt.Errorf("no content found")
	}

	return &GeminiCompleteAIResponse{
		Content:      candidate.Content.Parts[0].Text,
		FinishReason: candidate.FinishReason,
	}, nil
}
