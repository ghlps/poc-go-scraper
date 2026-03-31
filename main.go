package main

import (
	"context"
	"go-scraper/internal/config"
	"go-scraper/internal/db"
	"go-scraper/internal/models"
	"go-scraper/internal/scraper"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using system env vars")
	}

	ctx := context.Background()
	cfg := config.Load()

	responseData, err := scraper.Scrape()
	if err != nil {
		log.Fatalf("scrape failed: %v", err)
	}

	executionState := models.ExecutionState{
		ExecutionId: uuid.New().String(),
		Status:      "SUCCESS",
		RuCode:      responseData.RuCode,
		RunType:     "PRIMARY",
		Menu:        responseData,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(72 * time.Hour),
	}

	store, err := db.NewStore(ctx, cfg)
	if err != nil {
		log.Fatalf("create store failed: %v", err)
	}

	if err := store.Save(ctx, executionState); err != nil {
		log.Fatalf("db save failed: %v", err)
	}

	log.Println("saved to database successfully!")
}
