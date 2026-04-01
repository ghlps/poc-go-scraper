package scraper

import (
	"context"
	"fmt"
	"go-scraper/internal/config"
	"go-scraper/internal/db"
	"go-scraper/internal/models"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

type EventLambda struct {
	RuCode  string `json:"ruCode"`
	RunType string `json:"runType"`
}

type Scraper struct {
	store *db.Store
	cfg   *config.Config
}

func New(ctx context.Context, cfg *config.Config) (*Scraper, error) {
	store, err := db.NewStore(ctx, *cfg)
	if err != nil {
		return nil, err
	}
	return &Scraper{
		store: store,
		cfg:   cfg,
	}, nil
}

func (s *Scraper) Handle(ctx context.Context, event EventLambda) error {
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

	responseData, err := scrape(time.Now(), restaurant)
	if err != nil {
		scraperExecution.Status = models.ExecutionStatusFailed
		scraperExecution.Menu = nil
	}

	scraperExecution.Menu = &responseData
	scraperExecution.Status = models.ExecutionStatusSuccess

	store, err := db.NewStore(ctx, cfg)
	if err != nil {
		return fmt.Errorf("create store failed: %w", err)
	}

	if err := store.Save(ctx, scraperExecution); err != nil {
		return fmt.Errorf("db save failed: %w", err)
	}

	log.Println("Saved to database successfully")
	return nil
}
