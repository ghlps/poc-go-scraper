package models

type ResponseData struct {
	Date    string            `json:"date"    dynamodbav:"date"`
	ImgMenu *string           `json:"imgMenu" dynamodbav:"imgMenu"`
	RuName  string            `json:"ruName"  dynamodbav:"ruName"`
	RuUrl   string            `json:"ruUrl"   dynamodbav:"ruUrl"`
	RuCode  string            `json:"ruCode"  dynamodbav:"ruCode"`
	Served  []string          `json:"served"  dynamodbav:"served"`
	Meals   map[string][]Meal `json:"meals"   dynamodbav:"meals"`
}
