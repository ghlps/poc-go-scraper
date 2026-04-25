package scraper

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ghlps/poc-go-scraper/internal/config"
	"github.com/ghlps/poc-go-scraper/internal/db"
	"github.com/ghlps/poc-go-scraper/internal/models"
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

func (s *Scraper) Handle(ctx context.Context, event *EventLambda) (*models.Menu, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using system env vars")
	}

	restaurantCode, err := models.ParseRestaurantCode(event.RuCode)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	restaurant := models.NewRestaurant(restaurantCode)
	timeToScrape := time.Now().AddDate(0, 0, event.DateOffset)

	fmt.Println(restaurant)

	execution := models.ScraperExecution{
		ExecutionId:    uuid.New().String(),
		RestaurantCode: restaurantCode.String(),
		MenuDate:       timeToScrape.Format("02/01/2006"),
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(72 * time.Hour).Unix(),
		Menu: &models.Menu{
			Restaurant: &restaurant,
			Meals:      make(map[string][]models.Meal),
		},
	}

	return s.decider(ctx, &execution, timeToScrape)
}

func (s *Scraper) decider(ctx context.Context, execution *models.ScraperExecution, timeToScrape time.Time) (*models.Menu, error) {
	return s.runCheckup(ctx, execution, timeToScrape)
}
