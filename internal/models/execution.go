package models

import (
	"time"
)

type ScraperExecution struct {
	ExecutionId string        `json:"executionId"   dynamodbav:"execution_id,string"`
	Status      string        `json:"status"        dynamodbav:"status"`
	RunType     RunType       `json:"runType"       dynamodbav:"run_type"`
	Restaurant  Restaurant    `json:"ru"  dynamodbav:"ru"`
	Menu        *ResponseData `json:"menu,omitempty" dynamodbav:"menu,omitempty"`
	CreatedAt   time.Time     `json:"createdAt"     dynamodbav:"created_at"`
	ExpiresAt   time.Time     `json:"expiresAt"     dynamodbav:"expires_at"`
}
