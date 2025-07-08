package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
)

type processRequestBody struct {
	Prompt string `json:"prompt"`
}

type streamReturnType struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Done    bool        `json:"done"` // so the client doesnt keep listening indefinitely
}

func (handler *Handler) PostRequestHandler(c echo.Context) error {
	flusher, err := GetSSEFlusher(c)
	if err != nil {
		return c.JSON(500, streamReturnType{
			Message: "Request error. Error: " + err.Error(),
			Data:    nil,
		})
	}

	var body processRequestBody
	if err := c.Bind(&body); err != nil {
		jsonResponse, _ := json.Marshal(streamReturnType{
			Message: "Invalid request body. Error: " + err.Error(),
			Data:    nil,
			Done:    true,
		})
		fmt.Fprintf(c.Response(), "data: %s\n\n", jsonResponse)
		return nil
	}

	prompt := body.Prompt
	if len(prompt) > 500 {
		jsonResponse, _ := json.Marshal(streamReturnType{
			Message: "Max prompt length is 500 characters.",
			Data:    nil,
			Done:    true,
		})
		fmt.Fprintf(c.Response(), "data: %s\n\n", jsonResponse)
		return nil
	}

	dataStream, err := handler.HandleRequest(prompt)
	if err != nil {
		jsonResponse, _ := json.Marshal(streamReturnType{
			Message: "Server error. Error: " + err.Error(),
			Data:    nil,
			Done:    true,
		})
		fmt.Fprintf(c.Response(), "data: %s\n\n", jsonResponse)
		return nil
	}

	for response := range dataStream {
		responseStruct := streamReturnType{
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

	finalResponseStruct := streamReturnType{
		Message: "",
		Data:    nil,
		Done:    true,
	}

	jsonFinalResponse, _ := json.Marshal(finalResponseStruct)
	fmt.Fprintf(c.Response(), "data: %s\n\n", jsonFinalResponse)

	return nil
}
