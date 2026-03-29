package main

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type Request struct {
	Date string `json:"date"`
}

type Response struct {
	Date    string            `json:"date"`
	ImgMenu *string           `json:"imgMenu"`
	RuName  string            `json:"ruName"`
	RuUrl   string            `json:"ruUrl"`
	RuCode  string            `json:"ruCode"`
	Served  []string          `json:"served"`
	Meals   map[string][]Meal `json:"meals"`
}

type Meal struct {
	Name  string   `json:"name"`
	Icons []string `json:"icons"`
}

func main() {
	responseData, err := scrape(time.Now().AddDate(0, 0, 1))
	if err != nil {
		log.Printf(responseData.Date)
	}

	file, err := os.Create("response.json")
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	err = encoder.Encode(responseData)
	if err != nil {
		log.Fatalf("Failed to encode JSON: %v", err)
	}

	log.Println("Successfully saved response.json!")
}
