package main

import (
	"context"
	"fmt"
	"go-scraper/internal/config"
	"go-scraper/internal/db"
	"go-scraper/internal/models"
	"go-scraper/internal/scraper"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

type Event struct {
	RuCode  string `json:"ru_code"`
	RunType string `json:"run_type"`
}

func handler(ctx context.Context, event Event) error {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using system env vars")
	}

	runType := "PRIMARY"
	log.Printf("%+v", event)
	if event.RunType != "" {
		runType = event.RunType
	}

	cfg := config.Load()

	responseData, err := scraper.Scrape()
	if err != nil {
		return fmt.Errorf("scrape failed: %w", err)
	}

	executionState := models.ExecutionState{
		ExecutionId: uuid.New().String(),
		Status:      "SUCCESS",
		RuCode:      responseData.RuCode,
		RunType:     runType,
		Menu:        responseData,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(72 * time.Hour),
	}

	store, err := db.NewStore(ctx, cfg)
	if err != nil {
		return fmt.Errorf("create store failed: %w", err)
	}

	if err := store.Save(ctx, executionState); err != nil {
		return fmt.Errorf("db save failed: %w", err)
	}

	log.Println("saved to database successfully!")
	return nil
}
