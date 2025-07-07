package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	handlers "server/internal/handlers/ai"
	"server/internal/routing"
)

const frontendUrl = "http://localhost:5173"

func main() {
	e := echo.New()
	e.HideBanner = true // why is it even false by default
	handler, err := handlers.NewHandler()
	if err != nil {
		panic(err)
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{frontendUrl},
		AllowMethods:     []string{echo.GET, echo.POST, echo.OPTIONS},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))
	InitRoutes(e, handler)

	e.Logger.Fatal(e.Start(":1323"))
}

func InitRoutes(e *echo.Echo, handler *handlers.Handler) {
	routing.InitGetRoutes(e, handler)
}
