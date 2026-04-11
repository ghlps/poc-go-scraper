package scraper

import (
	"context"
	"fmt"
	"go-scraper/internal/models"
	"log"
	"time"
)

func (s *Scraper) runCheckup(ctx context.Context, execution models.ScraperExecution, timeToScrape time.Time) (*MenuDiff, error) {
	date := timeToScrape.Format("02/01/2006")
	MenuData, err := scrape(timeToScrape, *execution.Menu.Restaurant)
	if err != nil {
		execution.Status = models.ExecutionStatusFailed
		if saveErr := s.store.Save(ctx, execution); saveErr != nil {
			log.Printf("failed to save the failed checkup execution: %v", saveErr)
		}
		return nil, fmt.Errorf("scrape failed: %w", err)
	}

	currentHash, err := hashMenu(&MenuData)
	if err != nil {
		return nil, fmt.Errorf("hashing failed: %w", err)
	}

	previous, err := s.store.GetLatestByDate(ctx, date, execution.Menu.Restaurant.Code.String())
	if err != nil {
		return nil, fmt.Errorf("fetch preivous execution: %w", err)
	}

	if previous == nil {
		log.Printf("no previous execution found for %s, saving as first entry", date)
		MenuData.Restaurant = execution.Menu.Restaurant
		execution.Menu = &MenuData
		execution.MenuHash = currentHash
		execution.Status = models.ExecutionStatusSuccess
		if err := s.store.Save(ctx, execution); err != nil {
			return nil, fmt.Errorf("db save failed: %w", err)
		}
		return nil, nil
	}

	if previous.MenuHash == currentHash {
		log.Printf("menu unchanged for %s, skipping save", date)
		return nil, nil
	}

	MenuData.Restaurant = execution.Menu.Restaurant
	execution.Menu = &MenuData
	execution.MenuHash = currentHash
	execution.Status = models.ExecutionStatusSuccess
	if err := s.store.Save(ctx, execution); err != nil {
		return nil, fmt.Errorf("db save failed: %w", err)
	}

	return &MenuDiff{
		Previous: previous.Menu,
		Current:  &MenuData,
	}, nil

}

func (s *Scraper) runPrimary(ctx context.Context, execution models.ScraperExecution, timeToScrape time.Time) (*models.Menu, error) {
	return s.scrapeAndSave(ctx, execution, timeToScrape)
}

func (s *Scraper) runBackup(ctx context.Context, execution models.ScraperExecution, timeToScrape time.Time) (*models.Menu, error) {
	hasFailed, err := s.store.HasFailedExecutionForDate(ctx, timeToScrape.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("check failed execution: %w", err)
	}
	if !hasFailed {
		log.Println("no failed execution found, skipping backup run")
		return nil, nil
	}
	return s.scrapeAndSave(ctx, execution, timeToScrape)
}

func (s *Scraper) scrapeAndSave(ctx context.Context, execution models.ScraperExecution, timeToScrape time.Time) (*models.Menu, error) {
	MenuData, err := scrape(timeToScrape, *execution.Menu.Restaurant)
	if err != nil {
		execution.Status = models.ExecutionStatusFailed
		if saveErr := s.store.Save(ctx, execution); saveErr != nil {
			log.Printf("failed to save failed execution: %v", saveErr)
		}
		return nil, fmt.Errorf("scrape failed: %w", err)
	}

	menuHash, err := hashMenu(&MenuData)
	if err != nil {
		return nil, fmt.Errorf("hashing failed: %w", err)
	}

	MenuData.Restaurant = execution.Menu.Restaurant
	execution.Menu = &MenuData
	execution.MenuHash = menuHash
	execution.Status = models.ExecutionStatusSuccess

	if err := s.store.Save(ctx, execution); err != nil {
		return nil, fmt.Errorf("db save failed: %w", err)
	}
	return &MenuData, nil
}
