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

type RunResult struct {
	Menu    *models.Menu               `json:"menu"`
	Changes map[string]models.MealDiff `json:"changes,omitempty"`
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

func (s *Scraper) Handle(ctx context.Context, event *EventLambda) (*RunResult, error) {
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

	fmt.Println(restaurant)

	execution := models.ScraperExecution{
		ExecutionId:    uuid.New().String(),
		RunType:        rt,
		RestaurantCode: restaurantCode.String(),
		MenuDate:       timeToScrape.Format("02/01/2006"),
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(72 * time.Hour),
		Menu: &models.Menu{
			Restaurant: &restaurant,
			Meals:      make(map[string][]models.Meal),
		},
	}

	return s.decider(ctx, &execution, timeToScrape)
}

func (s *Scraper) decider(ctx context.Context, execution *models.ScraperExecution, timeToScrape time.Time) (*RunResult, error) {
	switch execution.RunType {
	case models.RunTypePrimary:
		menu, err := s.runPrimary(ctx, execution, timeToScrape)
		if err != nil {
			return nil, err
		}
		return &RunResult{Menu: menu}, nil

	case models.RunTypeBackup:
		menu, err := s.runBackup(ctx, execution, timeToScrape)
		if err != nil {
			return nil, err
		}
		return &RunResult{Menu: menu}, nil

	case models.RunTypeCheckup:
		result, err := s.runCheckup(ctx, execution, timeToScrape)
		if err != nil {
			return nil, err
		}

		changedMeals := make(map[string][]models.Meal)
		for mealType := range result.Changes {
			if meals, ok := result.Menu.Meals[mealType]; ok {
				changedMeals[mealType] = meals
			}
		}
		partialMenu := &models.Menu{
			Restaurant: result.Menu.Restaurant,
			Date:       result.Menu.Date,
			ImgMenu:    result.Menu.ImgMenu,
			Served:     result.Menu.Served,
			Meals:      changedMeals,
		}
		return &RunResult{Menu: partialMenu, Changes: result.Changes}, nil
	default:
		return nil, fmt.Errorf("unknown run type: %s", execution.RunType)
	}
}
