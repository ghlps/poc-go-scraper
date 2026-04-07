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

type RunResult struct {
	Menu *models.ResponseData // set for primary/backup runs
	Diff *MenuDiff            // set for checkup runs
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

func (s *Scraper) Handle(ctx context.Context, event EventLambda) (*RunResult, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using system env vars")
	}

	rt, err := models.ParseRunType(event.RunType)
	if err != nil {
		log.Printf("validation error: %v", err)
		return nil, fmt.Errorf("validation error: %w", err)
	}

	restaurantCode, err := models.ParseRestaurantCode(event.RuCode)
	if err != nil {
		log.Printf("validation error: %v", err)
		return nil, fmt.Errorf("validation error: %w", err)
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
		log.Printf("running a CHECKUP run")
		diff, err := s.runCheckup(ctx, execution, timeToScrape)
		if err != nil {
			return nil, err
		}
		return &RunResult{Diff: diff}, nil
	default:
		return nil, fmt.Errorf("unknown run type: %s", rt)
	}
}

type MenuDiff struct {
	Previous *models.ResponseData
	Current  *models.ResponseData
}

func (s *Scraper) runCheckup(ctx context.Context, execution models.ScraperExecution, timeToScrape time.Time) (*MenuDiff, error) {
	date := timeToScrape.Format("02/01/2006")
	responseData, err := scrape(timeToScrape, execution.Restaurant)
	if err != nil {
		execution.Status = models.ExecutionStatusFailed
		if saveErr := s.store.Save(ctx, execution); saveErr != nil {
			log.Printf("failed to save the failed checkup execution: %v", saveErr)
		}
		return nil, fmt.Errorf("scrape failed: %w", err)
	}

	currentHash, err := hashMenu(&responseData)
	if err != nil {
		return nil, fmt.Errorf("hashing failed: %w", err)
	}

	previous, err := s.store.GetLatestByDate(ctx, date, execution.Restaurant.Code.String())
	if err != nil {
		return nil, fmt.Errorf("fetch preivous execution: %w", err)
	}
	log.Printf("GetLatestByDate result for %s: %+v", date, previous) // ← add this

	if previous == nil {
		log.Printf("no previous execution found for %s, saving as first entry", date)
		execution.Menu = &responseData
		execution.MenuHash = currentHash
		execution.Status = models.ExecutionStatusSuccess
		if err := s.store.Save(ctx, execution); err != nil {
			return nil, fmt.Errorf("db save failed: %w", err)
		}
		return nil, nil
	}

	// Hashes match — nothing changed, skip saving
	if previous.MenuHash == currentHash {
		log.Printf("menu unchanged for %s, skipping save", date)
		return nil, nil
	}

	// Hashes differ — save new execution and return diff
	execution.Menu = &responseData
	execution.MenuHash = currentHash
	execution.Status = models.ExecutionStatusSuccess
	if err := s.store.Save(ctx, execution); err != nil {
		return nil, fmt.Errorf("db save failed: %w", err)
	}

	return &MenuDiff{
		Previous: previous.Menu,
		Current:  &responseData,
	}, nil

}

func (s *Scraper) runPrimary(ctx context.Context, execution models.ScraperExecution, timeToScrape time.Time) (*models.ResponseData, error) {
	return s.scrapeAndSave(ctx, execution, timeToScrape)
}

func (s *Scraper) runBackup(ctx context.Context, execution models.ScraperExecution, timeToScrape time.Time) (*models.ResponseData, error) {
	hasFailed, err := s.store.HasFailedExecutionForDate(ctx, timeToScrape.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("check failed execution: %w", err)
	}
	if !hasFailed {
		log.Println("no failed execution found, skipping backup run")
		return nil, nil // nil menu signals "skipped"
	}
	return s.scrapeAndSave(ctx, execution, timeToScrape)
}

func (s *Scraper) scrapeAndSave(ctx context.Context, execution models.ScraperExecution, timeToScrape time.Time) (*models.ResponseData, error) {
	responseData, err := scrape(timeToScrape, execution.Restaurant)
	if err != nil {
		execution.Status = models.ExecutionStatusFailed
		if saveErr := s.store.Save(ctx, execution); saveErr != nil {
			log.Printf("failed to save failed execution: %v", saveErr)
		}
		return nil, fmt.Errorf("scrape failed: %w", err)
	}

	menuHash, err := hashMenu(&responseData)
	if err != nil {
		return nil, fmt.Errorf("hashing failed: %w", err)
	}

	execution.Menu = &responseData
	execution.MenuHash = menuHash
	execution.Status = models.ExecutionStatusSuccess

	if err := s.store.Save(ctx, execution); err != nil {
		return nil, fmt.Errorf("db save failed: %w", err)
	}
	return &responseData, nil
}
