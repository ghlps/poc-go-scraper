package scraper

import (
	"fmt"
	"go-scraper/internal/models"
	"log"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

func Scrape() (models.ResponseData, error) {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36"),
	)

	c.SetRequestTimeout(15 * time.Second)

	c.OnRequest(func(r *colly.Request) {
		log.Printf("Visiting: %s", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		log.Println("Everything connected")
	})

	c.OnError(func(r *colly.Response, err error) {
		if r != nil {
			log.Printf("Request URL: %s failed with response: %v\nStatus Code: %d\nError: %v", r.Request.URL, r, r.StatusCode, err)
		} else {
			log.Printf("Request failed with error: %v", err)
		}
	})

	return transverseDOM(time.Now(), c)
}

func transverseDOM(dateScraped time.Time, c *colly.Collector) (models.ResponseData, error) {
	state := &scrapeState{
		payload: models.ResponseData{
			Date:   getFormattedDate(dateScraped),
			RuName: "JARDIM BOTÂNICO",
			RuUrl:  "https://pra.ufpr.br/ru/cardapio-ru-jardim-botanico/",
			RuCode: "BOT",
			Served: []string{"breakfast", "lunch", "dinner"},
			Meals:  make(map[string][]models.Meal),
		},
	}

	state.checkIfDateExists(c, getFormattedDate(dateScraped))
	state.formatEachRow(c)

	c.OnScraped(func(r *colly.Response) {
		state.saveMeals()
		log.Println("Scraping completed.")
	})

	if err := c.Visit(state.payload.RuUrl); err != nil {
		return models.ResponseData{}, fmt.Errorf("visit page: %w", err)
	}

	return state.payload, nil
}

func (s *scrapeState) checkIfDateExists(c *colly.Collector, formattedDate string) {
	c.OnHTML("strong", func(e *colly.HTMLElement) {
		if s.dateFound {
			return
		}
		dateText := strings.TrimSpace(e.Text)
		if strings.Contains(dateText, formattedDate) {
			log.Printf("Matching date found: %s", formattedDate)
			s.dateFound = true
			s.tableFound = false
		}
	})
}

func (s *scrapeState) formatEachRow(c *colly.Collector) {
	c.OnHTML("figure.wp-block-table", func(e *colly.HTMLElement) {
		if !s.dateFound || s.tableFound {
			return
		}
		s.tableFound = true

		e.ForEach("tr", func(_ int, row *colly.HTMLElement) {
			row.ForEach("td", func(_ int, cell *colly.HTMLElement) {
				s.processCell(cell)
			})
		})
	})
}

func (s *scrapeState) processCell(cell *colly.HTMLElement) {
	htmlContent := strings.ToUpper(strings.TrimSpace(cell.Text))

	if isMealType(htmlContent) {
		s.saveMeals()
		s.currentMealType = mapMealType(htmlContent)
		return
	}

	s.parseMealItems(cell)
}

func (s *scrapeState) parseMealItems(cell *colly.HTMLElement) {
	cellHTML, err := cell.DOM.Html()
	if err != nil {
		log.Printf("Error getting cell HTML: %v", err)
		return
	}

	for _, part := range strings.Split(cellHTML, "\n") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		meal, err := parseMeal(part)
		if err != nil || meal.Name == "" {
			continue
		}

		log.Printf("Adding meal: %s | icons: %v", meal.Name, meal.Icons)
		s.mealOptions = append(s.mealOptions, meal)
	}
}
