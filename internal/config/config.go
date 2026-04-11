package config

import (
	"os"
)

const (
	envDynamoURL = "DYNAMO_URL"
)

type Config struct {
	DynamoURL string
	IsDev     bool
}

func Load() Config {
	dynamoURL := getEnv(envDynamoURL, "")
	if dynamoURL == "" && os.Getenv("AWS_LAMBDA_FUNCTION_NAME") == "" {
		dynamoURL = os.Getenv("DYNAMO_URL")
	}
	return Config{
		DynamoURL: dynamoURL,
		IsDev:     os.Getenv("AWS_LAMBDA_FUNCTION_NAME") == "",
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
