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

	previous, err := s.store.GetLatestByDate(ctx, date, execution.Menu.Restaurant.Code.String())
	if err != nil {
		return nil, fmt.Errorf("fetch previous execution: %w", err)
	}

	menuData.Restaurant = execution.Menu.Restaurant

	if previous == nil {
		log.Printf("no previous execution found for %s, saving as first entry", date)
		execution.Menu = &menuData
		execution.MenuHash = currentHash
		execution.Status = models.ExecutionStatusSuccess
		if err := s.store.Save(ctx, *execution); err != nil {
			return nil, fmt.Errorf("db save failed: %w", err)
		}
		return &menuData, nil
	}

	if previous.MenuHash == currentHash {
		log.Printf("menu unchanged for %s, skipping save", date)
		return &menuData, nil
	}

	markChangedMeals(previous.Menu, &menuData)

	execution.Menu = &menuData
	execution.MenuHash = currentHash
	execution.Status = models.ExecutionStatusSuccess
	if err := s.store.Save(ctx, *execution); err != nil {
		return nil, fmt.Errorf("db save failed: %w", err)
	}

	return &menuData, nil
}
