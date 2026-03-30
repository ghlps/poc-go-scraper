package scraper

import (
	"go-scraper/internal/models"
	"log"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

func Scrape(dateToScrape time.Time) (models.ResponseData, error) {
	formattedDate := getFormattedDate(dateToScrape)
	log.Printf("Doing a request with the date %s", formattedDate)
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36"),
	)

	c.SetRequestTimeout(15 * time.Second)

	c.OnRequest(func(r *colly.Request) {
		log.Printf("Visiting: %s", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		log.Printf("Everything connected")
	})

	c.OnError(func(r *colly.Response, err error) {
		if r != nil {
			log.Printf("Request URL: %s failed with response: %v\nStatus Code: %d\nResponse Body: %s\nError: %v", r.Request.URL, r, r.StatusCode, string(r.Body), err)
		} else {
			log.Printf("Request failed with error: %v", err)
		}
	})

	return actuallyScrape(formattedDate, c)
}

func checkIfDateExists(c *colly.Collector, dateFound bool, formattedDate string, tableFound bool) {
	c.OnHTML("strong", func(e *colly.HTMLElement) {
		if !dateFound {
			dateText := strings.TrimSpace(e.Text)
			log.Printf("Found date: %s", dateText)

			if strings.Contains(dateText, formattedDate) {
				log.Printf("Matching date found: %s", formattedDate)
				dateFound = true
				tableFound = false
			}
		}
	})
}

func actuallyScrape(formattedDate string, c *colly.Collector) (models.ResponseData, error) {
	responsePayload := models.ResponseData{
		Date:    formattedDate,
		ImgMenu: nil,
		RuName:  "JARDIM BOTÂNICO",
		RuUrl:   "https://pra.ufpr.br/ru/cardapio-ru-jardim-botanico/",
		RuCode:  "BOT",
		Served:  []string{"breakfast", "lunch", "dinner"},
		Meals:   make(map[string][]models.Meal),
	}

	var currentMealType string
	var mealOptions []models.Meal

	var dateFound bool
	var tableFound bool

	log.Printf("Starting to scrape the page: %s", responsePayload.RuUrl)
	checkIfDateExists(c, dateFound, formattedDate, tableFound)

	c.OnHTML("figure.wp-block-table", func(e *colly.HTMLElement) {
		if dateFound && !tableFound {
			log.Println("Found the meal table. Starting extraction...")
			tableFound = true

			e.ForEach("tr", func(_ int, row *colly.HTMLElement) {
				row.ForEach("td", func(_ int, cell *colly.HTMLElement) {
					htmlContent := strings.ToUpper(strings.TrimSpace(cell.Text))

					if strings.Contains(htmlContent, "CAFÉ DA MANHÃ") ||
						strings.Contains(htmlContent, "ALMOÇO") ||
						strings.Contains(htmlContent, "JANTAR") {

						if len(mealOptions) > 0 {
							log.Printf("Saving meals for: %s", currentMealType)
							responsePayload.Meals[currentMealType] = mealOptions
							mealOptions = nil
						}

						currentMealType = mapMealType(htmlContent)
						log.Printf("Current meal type: %s", currentMealType)

					} else {
						cellHTML, err := cell.DOM.Html()
						if err != nil {
							log.Printf("Error getting cell HTML: %v", err)
							return
						}

						parts := strings.Split(cellHTML, "\n")
						for _, part := range parts {
							part = strings.TrimSpace(part)
							if part == "" {
								continue
							}

							partDOM, err := goquery.NewDocumentFromReader(strings.NewReader(part))
							if err != nil {
								log.Printf("Error parsing part: %v", err)
								continue
							}

							name := strings.TrimSpace(partDOM.Text())
							name = strings.Join(strings.Fields(name), " ")
							if name == "" {
								continue
							}

							icons := []string{}
							partDOM.Find("img").Each(func(_ int, img *goquery.Selection) {
								if title, exists := img.Attr("title"); exists && title != "" {
									icons = append(icons, title)
								}
							})

							log.Printf("Adding meal item: %s | icons: %v", name, icons)
							mealOptions = append(mealOptions, models.Meal{
								Name:  name,
								Icons: icons,
							})
						}
					}
				})
			})
		}
	})

	if len(mealOptions) > 0 {
		log.Printf("Saving meals for: %s", currentMealType)
		responsePayload.Meals[currentMealType] = mealOptions
	}

	c.OnScraped(func(r *colly.Response) {
		log.Println("Scraping completed.")
	})

	err := c.Visit("https://pra.ufpr.br/ru/cardapio-ru-jardim-botanico/")
	if err != nil {
		log.Printf("Error visiting page: %v", err)
		return models.ResponseData{}, err
	}

	if len(mealOptions) > 0 {
		responsePayload.Meals[currentMealType] = mealOptions
	}

	log.Println("Successfully scraped and created the response data.")
	return responsePayload, nil
}
