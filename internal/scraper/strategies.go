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
		return nil, fmt.Errorf("scrape failed: %w", err)
	}

	currentHash, err := hashMenu(&menuData)
	if err != nil {
		return nil, fmt.Errorf("hashing failed: %w", err)
	}

	lastExecution, err := s.store.GetLatestByDay(ctx, date, execution.Menu.Restaurant.Code.String())
	if err != nil {
		return nil, err
	}

	if lastExecution != nil {
		if lastExecution.MenuHash == currentHash {
			log.Printf("The menu didn't change %s, skipping...", date)
			return nil, nil
		}

		markChangedMeals(lastExecution.Menu, &menuData)
	}

	menuData.Restaurant = execution.Menu.Restaurant
	execution.Menu = &menuData
	execution.MenuHash = currentHash
	execution.Status = models.ExecutionStatusSuccess

	if err := s.store.Save(ctx, *execution); err != nil {
		return nil, fmt.Errorf("db save failed: %w", err)
	}

	return &menuData, nil
}
