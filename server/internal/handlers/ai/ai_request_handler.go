package handlers

import (
	"github.com/labstack/echo/v4"
	"server/internal/handlers"
	"time"
)

type processRequestBody struct {
	Prompt string `json:"prompt"`
}

func (h *Handler) GetRequestHandler(c echo.Context) error {
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
	response, err := h.HandleRequest(prompt)
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
