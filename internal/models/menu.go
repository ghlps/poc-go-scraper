package models

type Menu struct {
	Restaurant *Restaurant       `json:"restaurant" dynamodbav:"restaurant"`
	Date       string            `json:"date"       dynamodbav:"date"`
	ImgMenu    *string           `json:"imgMenu"    dynamodbav:"img_menu"`
	Served     []string          `json:"served"     dynamodbav:"served"`
	Meals      map[string][]Meal `json:"meals"      dynamodbav:"meals"`
}

type Meal struct {
	Name    string   `json:"name"  dynamodbav:"name"`
	Icons   []string `json:"icons" dynamodbav:"icons"`
	Changed bool     `json:"changed" dynamodbav:"changed"`
}
