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
	RuCode     string `json:"ruCode"`
	RunType    string `json:"runType"`
	DateOffset int    `json:"dateOffset"`
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
	timeToScrape := time.Now().AddDate(0, 0, event.DateOffset)
	execution := models.ScraperExecution{
		ExecutionId: uuid.New().String(),
		Restaurant:  restaurant,
		RunType:     rt,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(72 * time.Hour),
	}

	switch rt {
	case models.RunTypePrimary:
		return s.runPrimary(ctx, execution, timeToScrape)
	case models.RunTypeBackup:
		return s.runBackup(ctx, execution, timeToScrape)
	case models.RunTypeCheckup:
		return s.runCheckup(ctx, execution, timeToScrape)
	default:
		return fmt.Errorf("unknown run type: %s", rt)
	}
}

func (s *Scraper) runBackup(ctx context.Context, execution models.ScraperExecution, timeToScrape time.Time) error {
	return nil
}

func (s *Scraper) runCheckup(ctx context.Context, execution models.ScraperExecution, timeToScrape time.Time) error {
	return nil
}

func (s *Scraper) runPrimary(ctx context.Context, execution models.ScraperExecution, timeToScrape time.Time) error {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using system env vars")
	}

	responseData, err := scrape(timeToScrape, execution.Restaurant)
	if err != nil {
		execution.Status = models.ExecutionStatusFailed
	} else {
		menuHash, err := hashMenu(&responseData)
		if err != nil {
			return fmt.Errorf("hashing failed: %w", err)
		}
		execution.Menu = &responseData
		execution.MenuHash = menuHash
		execution.Status = models.ExecutionStatusSuccess
	}

	if err := s.store.Save(ctx, execution); err != nil {
		return fmt.Errorf("db save failed: %w", err)
	}

	log.Println("Saved to database successfully")
	return nil
}
