package models

type Menu struct {
	Restaurant *Restaurant       `json:"restaurant" dynamodbav:"restaurant"`
	Date       string            `json:"date"       dynamodbav:"date"`
	ImgMenu    *string           `json:"imgMenu"    dynamodbav:"img_menu"`
	Served     []string          `json:"served"     dynamodbav:"served"`
	Meals      map[string][]Meal `json:"meals"      dynamodbav:"meals"`
}

type Meal struct {
	Name  string   `json:"name"  dynamodbav:"name"`
	Icons []string `json:"icons" dynamodbav:"icons"`
}

type MealDiff struct {
	Added   []Meal `json:"added"`
	Removed []Meal `json:"removed"`
}

type MenuCheckupResult struct {
	Menu    *Menu               `json:"menu"`
	Changes map[string]MealDiff `json:"changes"`
}
