package models

type Meal struct {
	Name  string   `json:"name"  dynamodbav:"name"`
	Icons []string `json:"icons" dynamodbav:"icons"`
}
