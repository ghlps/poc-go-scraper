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

type MenuDiff struct {
	Previous *models.Menu
	Current  *models.Menu
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

func (s *Scraper) Handle(ctx context.Context, event EventLambda) (*models.Menu, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using system env vars")
	}

	rt, err := models.ParseRunType(event.RunType)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	restaurantCode, err := models.ParseRestaurantCode(event.RuCode)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	restaurant := models.NewRestaurant(restaurantCode)
	timeToScrape := time.Now().AddDate(0, 0, event.DateOffset)

	execution := models.ScraperExecution{
		ExecutionId: uuid.New().String(),
		RunType:     rt,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(72 * time.Hour),
		Menu: &models.Menu{
			Restaurant: &restaurant,
			Meals:      make(map[string][]models.Meal),
		},
	}

	return s.decider(ctx, execution, timeToScrape)
}

func (s *Scraper) decider(ctx context.Context, execution models.ScraperExecution, timeToScrape time.Time) (*models.Menu, error) {
	switch execution.RunType {
	case models.RunTypePrimary:
		menu, err := s.runPrimary(ctx, execution, timeToScrape)
		if err != nil {
			return nil, err
		}
		return menu, nil

	case models.RunTypeBackup:
		menu, err := s.runBackup(ctx, execution, timeToScrape)
		if err != nil {
			return nil, err
		}
		return menu, nil

	case models.RunTypeCheckup:
		_, err := s.runCheckup(ctx, execution, timeToScrape)
		if err != nil {
			return nil, err
		}
		return execution.Menu, nil

	default:
		return nil, fmt.Errorf("unknown run type: %s", execution.RunType)
	}
}
