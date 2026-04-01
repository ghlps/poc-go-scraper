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
	RuCode  string `json:"ruCode"`
	RunType string `json:"runType"`
}

func handler(ctx context.Context, event Event) error {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using system env vars")
	}

	rt, err := models.ParseRunType(event.RunType)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	restaurantCode, err := models.ParseRestaurantCode(event.RuCode)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	restaurant := models.NewRestaurant(restaurantCode)

	cfg := config.Load()

	scraperExecution := models.ScraperExecution{
		ExecutionId: uuid.New().String(),
		Restaurant:  restaurant,
		RunType:     rt,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(72 * time.Hour),
	}

	responseData, err := scraper.Scrape(time.Now(), restaurant)
	if err != nil {
		scraperExecution.Status = "FAIL"
		scraperExecution.Menu = nil
	}

	scraperExecution.Menu = &responseData
	scraperExecution.Status = "SUCCESS"

	store, err := db.NewStore(ctx, cfg)
	if err != nil {
		return fmt.Errorf("create store failed: %w", err)
	}

	if err := store.Save(ctx, scraperExecution); err != nil {
		return fmt.Errorf("db save failed: %w", err)
	}

	log.Println("saved to database successfully!")
	return nil
}
