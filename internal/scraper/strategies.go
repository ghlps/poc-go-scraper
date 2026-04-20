package scraper

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ghlps/poc-go-scraper/internal/models"
)

func (s *Scraper) runCheckup(ctx context.Context, execution *models.ScraperExecution, timeToScrape time.Time) (*models.Menu, error) {
	date := timeToScrape.Format("02/01/2006")
	menuData, err := scrape(timeToScrape, *execution.Menu.Restaurant)
	if err != nil {
		execution.Status = models.ExecutionStatusFailed
		if saveErr := s.store.Save(ctx, *execution); saveErr != nil {
			log.Printf("failed to save the failed checkup execution: %v", saveErr)
		}
		return nil, fmt.Errorf("scrape failed: %w", err)
	}

	currentHash, err := hashMenu(&menuData)
	if err != nil {
		return nil, fmt.Errorf("hashing failed: %w", err)
	}

	menuData.Restaurant = execution.Menu.Restaurant

	existingWithHash, err := s.store.GetLatestByHash(ctx, date, execution.Menu.Restaurant.Code.String(), currentHash)
	if err != nil {
		return nil, fmt.Errorf("Fetch previous execution: %w", err)
	}

	if existingWithHash != nil {
		log.Printf("Execution with same hash already exists for %s, skipping save", date)
		return nil, nil
	}

	execution.Menu = &menuData
	execution.MenuHash = currentHash
	execution.Status = models.ExecutionStatusSuccess
	if err := s.store.Save(ctx, *execution); err != nil {
		return nil, fmt.Errorf("db save failed: %w", err)
	}

	log.Printf("New menu saved for %s with hash %s", date, currentHash)
	return &menuData, nil
}
