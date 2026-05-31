package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port    string
	GinMode string

	DBPath string // SQLite database file path

	JWTSecret        string
	JWTExpiry        time.Duration
	JWTRefreshExpiry time.Duration

	GeminiAPIKey      string
	AITimeoutSeconds  int
	WAYNXBaseURL      string // WAYNX recommendation API base URL

	GoogleClientID     string
	GoogleClientSecret string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	return &Config{
		Port:    getEnv("PORT", "8080"),
		GinMode: getEnv("GIN_MODE", "debug"),

		DBPath: getEnv("DB_PATH", "./kemit.db"),

		JWTSecret:        getEnv("JWT_SECRET", "dev-secret"),
		JWTExpiry:        parseDuration(getEnv("JWT_EXPIRY", "24h")),
		JWTRefreshExpiry: parseDuration(getEnv("JWT_REFRESH_EXPIRY", "168h")),

		GeminiAPIKey:      getEnv("GEMINI_API_KEY", ""),
		AITimeoutSeconds:  getEnvInt("AI_TIMEOUT_SECONDS", 30),
		WAYNXBaseURL:      getEnv("WAYNX_API_URL", "https://waynx-api-production.up.railway.app"),

		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}
