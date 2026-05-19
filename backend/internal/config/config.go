package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port            string
	FrontendBaseURL string
	NetplayBaseURL  string
	StorePath       string
	HTTPTimeout     time.Duration
}

func Load() Config {
	timeoutSeconds := getInt("HTTP_TIMEOUT_SECONDS", 10)

	return Config{
		Port:            getEnv("PORT", "8080"),
		FrontendBaseURL: getEnv("FRONTEND_BASE_URL", "http://localhost:5173"),
		NetplayBaseURL:  getEnv("NETPLAY_BASE_URL", "https://ctazh5lrhe.execute-api.ap-southeast-3.amazonaws.com/dev/api"),
		StorePath:       getEnv("STORE_PATH", "data/activations.json"),
		HTTPTimeout:     time.Duration(timeoutSeconds) * time.Second,
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
