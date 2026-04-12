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
	MenuData, err := scrape(timeToScrape, *execution.Menu.Restaurant)
	if err != nil {
		execution.Status = models.ExecutionStatusFailed
		if saveErr := s.store.Save(ctx, *execution); saveErr != nil {
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
		if err := s.store.Save(ctx, *execution); err != nil {
			return nil, fmt.Errorf("db save failed: %w", err)
		}
		return &MenuData, nil
	}

	if previous.MenuHash == currentHash {
		log.Printf("menu unchanged for %s, skipping save", date)
		return &MenuData, nil
	}

	// Menu changed - return only the changed meal types
	changedMenu := &models.Menu{
		Restaurant: execution.Menu.Restaurant,
		Meals:      make(map[string][]models.Meal),
	}

	for mealType, currentMeals := range MenuData.Meals {
		previousMeals, exists := previous.Menu.Meals[mealType]

		// If meal type didn't exist before or the meals are different, include it
		if !exists {
			changedMenu.Meals[mealType] = currentMeals
			log.Printf("Detected NEW meal type: %s with %d meals", mealType, len(currentMeals))
		} else if !mealsAreEqual(previousMeals, currentMeals) {
			changedMenu.Meals[mealType] = currentMeals
			log.Printf("Detected CHANGED meal type: %s (prev: %d meals, curr: %d meals)", mealType, len(previousMeals), len(currentMeals))
		}
	}

	// Also check for meal types that were removed
	for mealType, prevMeals := range previous.Menu.Meals {
		if _, exists := MenuData.Meals[mealType]; !exists {
			log.Printf("Detected REMOVED meal type: %s (was %d meals)", mealType, len(prevMeals))
		}
	}

	log.Printf("Changed meals to return: %d meal types", len(changedMenu.Meals))

	MenuData.Restaurant = execution.Menu.Restaurant
	execution.Menu = &MenuData
	execution.MenuHash = currentHash
	execution.Status = models.ExecutionStatusSuccess
	if err := s.store.Save(ctx, *execution); err != nil {
		return nil, fmt.Errorf("db save failed: %w", err)
	}

	return changedMenu, nil

}

func mealsAreEqual(meals1, meals2 []models.Meal) bool {
	if len(meals1) != len(meals2) {
		return false
	}

	for i, meal := range meals1 {
		if meal.Name != meals2[i].Name {
			return false
		}
		if len(meal.Icons) != len(meals2[i].Icons) {
			return false
		}
		for j, icon := range meal.Icons {
			if icon != meals2[i].Icons[j] {
				return false
			}
		}
	}

	return true
}

func (s *Scraper) runPrimary(ctx context.Context, execution *models.ScraperExecution, timeToScrape time.Time) (*models.Menu, error) {
	return s.scrapeAndSave(ctx, execution, timeToScrape)
}

func (s *Scraper) runBackup(ctx context.Context, execution *models.ScraperExecution, timeToScrape time.Time) (*models.Menu, error) {
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

func (s *Scraper) scrapeAndSave(ctx context.Context, execution *models.ScraperExecution, timeToScrape time.Time) (*models.Menu, error) {
	MenuData, err := scrape(timeToScrape, *execution.Menu.Restaurant)
	if err != nil {
		execution.Status = models.ExecutionStatusFailed
		if saveErr := s.store.Save(ctx, *execution); saveErr != nil {
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

	if err := s.store.Save(ctx, *execution); err != nil {
		return nil, fmt.Errorf("db save failed: %w", err)
	}
	return &MenuData, nil
}
