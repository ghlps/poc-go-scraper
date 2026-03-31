//go:build dev

package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
)

func main() {
	data, err := os.ReadFile("event.json")
	if err != nil {
		log.Fatalf("failed to read event.json: %v", err)
	}

	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		log.Fatalf("failed to parse event.json: %v", err)
	}

	if err := handler(context.Background(), event); err != nil {
		log.Fatal(err)
	}
}
