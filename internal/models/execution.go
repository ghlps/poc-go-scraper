package models

import (
	"time"
)

type ScraperExecution struct {
	ExecutionId string          `json:"executionId"   dynamodbav:"execution_id,string"`
	Status      ExecutionStatus `json:"status"        dynamodbav:"status"`
	RunType     RunType         `json:"runType"       dynamodbav:"run_type"`
	MenuHash    string          `json:"menuHash,omitempty" dynamodbav:"menu_hash,omitempty"`
	Menu        *Menu           `json:"menu,omitempty" dynamodbav:"menu,omitempty"`
	CreatedAt   time.Time       `json:"createdAt"     dynamodbav:"created_at"`
	ExpiresAt   time.Time       `json:"expiresAt"     dynamodbav:"expires_at"`
}
