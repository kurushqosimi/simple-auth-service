package main

import (
	"context"
	"fullstack-simple-app/config"
	"fullstack-simple-app/internal/app"
	"log"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	a, err := app.New(cfg)
	if err != nil {
		log.Fatalf("app.NewApp: %v", err)
	}

	err = a.Run(ctx)
	if err != nil {
		log.Fatalf("app.Run: %v", err)
	}
}
