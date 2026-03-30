package main

import (
	"context"
	"go-scraper/internal/db"
	"go-scraper/internal/scraper"
	"log"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()
	_ = godotenv.Load()

	responseData, err := scraper.Scrape(time.Now())
	if err != nil {
		log.Fatalf("scrape failed: %v", err)
	}

	store, err := db.NewStore(ctx)
	if err != nil {
		log.Fatalf("create store failed: %v", err)
	}

	if err := store.Save(ctx, responseData); err != nil {
		log.Fatalf("db save failed: %v", err)
	}

	log.Println("saved to file and db!")
}
