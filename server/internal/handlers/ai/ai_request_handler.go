package handlers

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"server/internal/handlers"
	"time"
)

type processRequestBody struct {
	Prompt string `json:"prompt"`
}

func (handler *Handler) PostRequestHandler(c echo.Context) error {
	var body processRequestBody
	if err := c.Bind(&body); err != nil {
		return c.JSON(400, handlers.ReturnType{
			Message: "Invalid request body. Error: " + err.Error(),
			Data:    nil,
		})
	}

	prompt := body.Prompt
	if len(prompt) > 500 {
		return c.JSON(400, handlers.ReturnType{
			Message: "Max prompt length is 500 characters.",
			Data:    nil,
		})
	}
	timer := time.Now()
	fmt.Println(prompt)
	response, err := handler.HandleRequest(prompt)
	if err != nil {
		return c.JSON(500, handlers.ReturnType{
			Message: "Server error. Error: " + err.Error(),
			Data:    nil,
		})
	}

	return c.JSON(200, handlers.ReturnType{
		Message: time.Since(timer).String(),
		Data:    response,
	})
}
