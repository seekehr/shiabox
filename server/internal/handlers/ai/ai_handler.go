package handlers

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"server/internal/llms"
)

// AIHandler - Handles the entire User -> AI communication.
type AIHandler struct {
	llms.Handler
}

// HandleRequest - Handle the entire prompt -> AI process, and return the SSE stream of tokens

// HandleCompletePrompt - For non-streaming uses
/*func (handler *AIHandler) HandleCompletePrompt(sysPrompt string, prompt string, model groq.Model) (*groq.CompleteAIResponse, error) {
	resp, err := groq.SendPrompt(sysPrompt, prompt, model, handler.llmApiKey, false)
	if err != nil {
		return nil, err
	}
	fmt.Println("Status code " + strconv.Itoa(resp.StatusCode))
	defer resp.Body.Close()
	return groq.ParseResponse(resp.Body)
}*/

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
