package routing

import (
	"github.com/labstack/echo/v4"
	handlers "server/internal/handlers/ai"
)

func InitGetRoutes(e *echo.Echo, handler *handlers.Handler) {
	ai := e.Group("ai")
	ai.POST("/request", handler.PostRequestHandler)
}
