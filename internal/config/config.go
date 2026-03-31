package config

import (
	"os"
)

const (
	envAppEnv    = "APP_ENV"
	envDynamoURL = "DYNAMO_URL"
)

type Config struct {
	AppEnv    string
	DynamoURL string
	IsDev     bool
}

func Load() Config {
	appEnv := getEnv(envAppEnv, "prod")
	return Config{
		AppEnv:    appEnv,
		DynamoURL: getEnv(envDynamoURL, ""),
		IsDev:     appEnv == "dev",
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
