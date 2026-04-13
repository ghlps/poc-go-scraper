package scraper

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ghlps/poc-go-scraper/internal/models"
)

func (s *Scraper) runCheckup(ctx context.Context, execution *models.ScraperExecution, timeToScrape time.Time) (*models.MenuCheckupResult, error) {
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
		return nil, fmt.Errorf("fetch previous execution: %w", err)
	}

	MenuData.Restaurant = execution.Menu.Restaurant

	if previous == nil {
		log.Printf("no previous execution found for %s, saving as first entry", date)
		execution.Menu = &MenuData
		execution.MenuHash = currentHash
		execution.Status = models.ExecutionStatusSuccess
		if err := s.store.Save(ctx, *execution); err != nil {
			return nil, fmt.Errorf("db save failed: %w", err)
		}
		return &models.MenuCheckupResult{Menu: &MenuData}, nil
	}

	if previous.MenuHash == currentHash {
		log.Printf("menu unchanged for %s, skipping save", date)
		return &models.MenuCheckupResult{Menu: &MenuData}, nil
	}

	changes := diffMenus(previous.Menu, &MenuData)

	execution.Menu = &MenuData
	execution.MenuHash = currentHash
	execution.Status = models.ExecutionStatusSuccess
	if err := s.store.Save(ctx, *execution); err != nil {
		return nil, fmt.Errorf("db save failed: %w", err)
	}

	return &models.MenuCheckupResult{Menu: &MenuData, Changes: changes}, nil
}

func diffMenus(previous, current *models.Menu) map[string]models.MealDiff {
	changes := make(map[string]models.MealDiff)

	for mealType, currentMeals := range current.Meals {
		previousMeals, existed := previous.Meals[mealType]
		if !existed {
			changes[mealType] = models.MealDiff{Added: currentMeals}
			log.Printf("Detected NEW meal type: %s with %d meals", mealType, len(currentMeals))
			continue
		}
		added, removed := diffMealSlices(previousMeals, currentMeals)
		if len(added) > 0 || len(removed) > 0 {
			changes[mealType] = models.MealDiff{Added: added, Removed: removed}
			log.Printf("Detected CHANGED meal type: %s (+%d, -%d)", mealType, len(added), len(removed))
		}
	}

	for mealType, previousMeals := range previous.Meals {
		if _, exists := current.Meals[mealType]; !exists {
			changes[mealType] = models.MealDiff{Removed: previousMeals}
			log.Printf("Detected REMOVED meal type: %s (was %d meals)", mealType, len(previousMeals))
		}
	}

	return changes
}

func diffMealSlices(previous, current []models.Meal) (added, removed []models.Meal) {
	prevIdx := indexMeals(previous)
	currIdx := indexMeals(current)

	for name, meal := range currIdx {
		if _, existed := prevIdx[name]; !existed {
			added = append(added, meal)
		}
	}
	for name, meal := range prevIdx {
		if _, exists := currIdx[name]; !exists {
			removed = append(removed, meal)
		}
	}
	return
}

func indexMeals(meals []models.Meal) map[string]models.Meal {
	idx := make(map[string]models.Meal, len(meals))
	for _, m := range meals {
		idx[m.Name] = m
	}
	return idx
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
