package config

import (
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	YoutubeAPIKeys    []string
	RateLimitRequests int
	RateLimitWindow   time.Duration
	TempDir           string
	ServerPort        string
}

var (
	instance *Config
	once     sync.Once
)

func Load() *Config {
	once.Do(func() {
		// Load .env file if it exists
		if err := godotenv.Load(); err != nil {
			// Try loading from env file as fallback
			godotenv.Load("env")
		}

		instance = &Config{
			YoutubeAPIKeys:    loadAPIKeys(),
			RateLimitRequests: getEnvInt("RATE_LIMIT_REQUESTS", 100),
			RateLimitWindow:   time.Duration(getEnvInt("RATE_LIMIT_WINDOW", 60)) * time.Second,
			TempDir:           getEnvString("TEMP_DIR", "./tmp"),
			ServerPort:        getEnvString("SERVER_PORT", "8080"),
		}

		validateConfig(instance)
	})
	return instance
}

func loadAPIKeys() []string {
	keys := []string{}

	// Main API key
	if mainKey := os.Getenv("YOUTUBE_API_KEY_MAIN"); mainKey != "" {
		keys = append(keys, mainKey)
	}

	// Backup keys
	for i := 1; i <= 5; i++ {
		key := os.Getenv("YOUTUBE_API_KEY_BACKUP_" + strconv.Itoa(i))
		if key != "" {
			keys = append(keys, key)
		}
	}

	return keys
}

func getEnvString(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func validateConfig(cfg *Config) {
	if len(cfg.YoutubeAPIKeys) == 0 {
		log.Fatal("Youtube API keys didn't emerge in environment variables")
	}

	if cfg.RateLimitRequests <= 0 {
		log.Fatal("Rate limit requests must be positive")
	}
}
