package main

import (
	"log"

	"ott-integration/backend/internal/app"
	"ott-integration/backend/internal/config"
	"ott-integration/backend/internal/handler"
)

func main() {
	cfg := config.Load()
	svc, err := app.NewService(cfg)
	if err != nil {
		log.Fatal(err)
	}

	router := handler.NewRouterFromConfig(svc, cfg)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
