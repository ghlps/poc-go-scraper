package scraper

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ghlps/poc-go-scraper/internal/models"
)

func (s *Scraper) scrapeAndSave(ctx context.Context, execution *models.ScraperExecution, timeToScrape time.Time) (*models.Menu, error) {
	menuData, err := scrape(timeToScrape, *execution.Menu.Restaurant)
	if err != nil {
		execution.Status = models.ExecutionStatusFailed
		if saveErr := s.store.Save(ctx, *execution); saveErr != nil {
			log.Printf("failed to save failed execution: %v", saveErr)
		}
		return nil, fmt.Errorf("scrape failed: %w", err)
	}

	menuHash, err := hashMenu(&menuData)
	if err != nil {
		return nil, fmt.Errorf("hashing failed: %w", err)
	}

	menuData.Restaurant = execution.Menu.Restaurant
	execution.Menu = &menuData
	execution.MenuHash = menuHash
	execution.Status = models.ExecutionStatusSuccess

	if err := s.store.Save(ctx, *execution); err != nil {
		return nil, fmt.Errorf("db save failed: %w", err)
	}
	return &menuData, nil
}

func indexMeals(meals []models.Meal) map[string]models.Meal {
	idx := make(map[string]models.Meal, len(meals))
	for _, m := range meals {
		idx[m.Name] = m
	}
	return idx
}

func markChangedMeals(previous, current *models.Menu) {
	for mealType, currentMeals := range current.Meals {
		previousMeals, existed := previous.Meals[mealType]
		if !existed {
			for i := range currentMeals {
				currentMeals[i].Changed = true
			}
			current.Meals[mealType] = currentMeals
			log.Printf("Detected NEW meal type: %s with %d meals", mealType, len(currentMeals))
			continue
		}

		prevIdx := indexMeals(previousMeals)
		changed := false
		for i, meal := range currentMeals {
			if _, existed := prevIdx[meal.Name]; !existed {
				currentMeals[i].Changed = true
				changed = true
			}
		}
		if changed {
			current.Meals[mealType] = currentMeals
			log.Printf("Detected CHANGED meal type: %s", mealType)
		}
	}

	for mealType := range previous.Meals {
		if _, exists := current.Meals[mealType]; !exists {
			log.Printf("Detected REMOVED meal type: %s", mealType)
		}
	}
}
