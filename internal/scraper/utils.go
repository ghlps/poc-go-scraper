package scraper

import (
	"strings"
	"time"
)

func getFormattedDate(date time.Time) string {
	return date.Format("02/01/2006")
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
