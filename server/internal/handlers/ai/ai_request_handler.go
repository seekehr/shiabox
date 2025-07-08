package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"server/internal/handlers"
)

type processRequestBody struct {
	Prompt string `json:"prompt"`
}

func (handler *Handler) PostRequestHandler(c echo.Context) error {
	flusher, err := GetSSEFlusher(c)
	if err != nil {
		return c.JSON(500, handlers.StreamReturnType{
			Message: "Request error. Error: " + err.Error(),
			Data:    nil,
		})
	}

	var body processRequestBody
	if err := c.Bind(&body); err != nil {
		jsonResponse, _ := json.Marshal(handlers.StreamReturnType{
			Message: "Invalid request body. Error: " + err.Error(),
			Data:    nil,
			Done:    true,
		})
		fmt.Fprintf(c.Response(), "data: %s\n\n", jsonResponse)
		return nil
	}

	prompt := body.Prompt
	if len(prompt) > 500 {
		jsonResponse, _ := json.Marshal(handlers.StreamReturnType{
			Message: "Max prompt length is 500 characters.",
			Data:    nil,
			Done:    true,
		})
		fmt.Fprintf(c.Response(), "data: %s\n\n", jsonResponse)
		return nil
	}

	dataStream, err := handler.HandleRequest(prompt)
	if err != nil {
		message := "Error handling request. Error: " + err.Error()
		if err.Error() == "ratelimit" {
			message = "30 msgs/s rate-limit reached of server. Please donate to help increase our rate-limit."
		}

		jsonResponse, _ := json.Marshal(handlers.StreamReturnType{
			Message: message,
			Data:    nil,
			Done:    true,
		})
		fmt.Fprintf(c.Response(), "data: %s\n\n", jsonResponse)
		return nil
	}

	for response := range dataStream {
		responseStruct := handlers.StreamReturnType{
			Message: "",
			Data:    response,
			Done:    false,
		}
		jsonResponse, _ := json.Marshal(responseStruct)
		// Write the given data to the response buffer as it is sent. Reason we need dis is cuz we might not send one msg per request and sometimes multiple
		fmt.Fprintf(c.Response(), "data: %s\n\n", jsonResponse)
		// Flush the buffer to send the response we have received immediately
		flusher.Flush()
	}

	finalResponseStruct := handlers.StreamReturnType{
		Message: "",
		Data:    nil,
		Done:    true,
	}

	jsonFinalResponse, _ := json.Marshal(finalResponseStruct)
	fmt.Fprintf(c.Response(), "data: %s\n\n", jsonFinalResponse)

	return nil
}
