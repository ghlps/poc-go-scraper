package scraper

import (
	"go-scraper/internal/models"
	"log"
)

type scrapeState struct {
	dateFound       bool
	tableFound      bool
	currentMealType string
	mealOptions     []models.Meal
	payload         models.ResponseData
}

func (s *scrapeState) saveMeals() {
	if len(s.mealOptions) > 0 {
		log.Printf("Saving meals for: %s", s.currentMealType)
		s.payload.Meals[s.currentMealType] = s.mealOptions
		s.mealOptions = nil
	}
}
