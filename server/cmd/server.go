package main

import (
	"github.com/labstack/echo/v4"
	handlers "server/internal/handlers/ai"
	"server/internal/routing"
)

func main() {
	e := echo.New()
	e.HideBanner = true // why is it even false by default
	handler, err := handlers.NewHandler()
	if err != nil {
		panic(err)
	}
	InitRoutes(e, handler)

	e.Logger.Fatal(e.Start(":1323"))
}

func InitRoutes(e *echo.Echo, handler *handlers.Handler) {
	routing.InitGetRoutes(e, handler)
}
