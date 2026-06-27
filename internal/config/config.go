package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppAddr        string
	CacheTTL       time.Duration
	MonobankAPIURL string
}

func Load() (Config, error) {
	ttlSeconds, err := getInt("CACHE_TTL_SECONDS", 300)
	if err != nil {
		return Config{}, err
	}

	if ttlSeconds <= 0 {
		return Config{}, fmt.Errorf("CACHE_TTL_SECONDS must be greater than zero")
	}

	return Config{
		AppAddr:        getString("APP_ADDR", ":8080"),
		CacheTTL:       time.Duration(ttlSeconds) * time.Second,
		MonobankAPIURL: getString("MONOBANK_API_URL", "https://api.monobank.ua/bank/currency"),
	}, nil
}

func getString(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func getInt(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}

	return parsed, nil
}
