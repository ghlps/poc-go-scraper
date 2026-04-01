package main

import (
	"context"
	"go-scraper/internal/config"
	"go-scraper/internal/scraper"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	svc, err := scraper.New(ctx, &cfg)
	if err != nil {
		log.Fatalf("failed to initialize scraper: %v", err)
	}

	lambda.Start(svc.Handle)
}
