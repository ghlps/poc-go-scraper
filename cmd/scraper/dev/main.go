package main

import (
	"context"
	"encoding/json"
	"go-scraper/internal/config"
	"go-scraper/internal/scraper"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system envs")
	}

	cfg := config.Load()
	ctx := context.Background()

	svc, err := scraper.New(ctx, &cfg)
	if err != nil {
		log.Fatalf("failed to initialize scraper service: %v", err)
	}

	data, err := os.ReadFile("event.json")
	if err != nil {
		log.Fatalf("failed to read event.json: %v", err)
	}

	var event scraper.EventLambda
	if err := json.Unmarshal(data, &event); err != nil {
		log.Fatalf("failed to parse event.json: %v", err)
	}

	log.Printf("Starting local scrape for: %s", event.RuCode)
	if err := svc.Handle(ctx, event); err != nil {
		log.Fatalf("Execution failed: %v", err)
	}

	log.Println("Local execution finished successfully")
}
