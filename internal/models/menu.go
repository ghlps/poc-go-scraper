package models

type ResponseData struct {
	Date    string            `json:"date"    dynamodbav:"date"`
	ImgMenu *string           `json:"imgMenu" dynamodbav:"img_menu"`
	Served  []string          `json:"served"  dynamodbav:"served"`
	Meals   map[string][]Meal `json:"meals"   dynamodbav:"meals"`
}
