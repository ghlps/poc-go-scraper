package scraper

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go-scraper/internal/models"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func parseMeal(part string) (models.Meal, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(part))
	if err != nil {
		return models.Meal{}, err
	}

	name := strings.Join(strings.Fields(doc.Text()), " ")

	icons := []string{}
	doc.Find("img").Each(func(_ int, img *goquery.Selection) {
		if title, exists := img.Attr("title"); exists && title != "" {
			icons = append(icons, title)
		}
	})

	return models.Meal{Name: name, Icons: icons}, nil
}

func hashMenu(menu *models.ResponseData) (string, error) {
	if menu == nil {
		return "", nil
	}
	b, err := json.Marshal(menu)
	if err != nil {
		return "", fmt.Errorf("hash marshal failed: %w", err)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func getFormattedDate(date time.Time) string {
	return date.Format("02/01/2006")
}

func isMealType(s string) bool {
	return strings.Contains(s, "CAFÉ DA MANHÃ") ||
		strings.Contains(s, "ALMOÇO") ||
		strings.Contains(s, "JANTAR")
}

func mapMealType(htmlContent string) string {
	if strings.Contains(htmlContent, "CAFÉ DA MANHÃ") {
		return "breakfast"
	} else if strings.Contains(htmlContent, "ALMOÇO") {
		return "lunch"
	} else if strings.Contains(htmlContent, "JANTAR") {
		return "dinner"
	}
	return ""
}
