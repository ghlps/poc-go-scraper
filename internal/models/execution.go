package models

import (
	"time"
)

type ExecutionState struct {
	ExecutionId string       `json:"executionId"   dynamodbav:"executionId,string"`
	Status      string       `json:"status"        dynamodbav:"status"`
	RuCode      string       `json:"ruCode"        dynamodbav:"ruCode"`
	RunType     string       `json:"runType"       dynamodbav:"runType"`
	Menu        ResponseData `json:"menu"          dynamodbav:"menu"`
	CreatedAt   time.Time    `json:"createdAt"     dynamodbav:"createdAt"`
	ExpiresAt   time.Time    `json:"expiresAt"     dynamodbav:"expiresAt"`
}
