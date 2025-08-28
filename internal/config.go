
package internal

import (
    "log"
    "os"
)

type Config struct {
    DatabaseDSN string
    JWTSecret   string
    EmailAPIKey string
    EmailFromAddress string
}

func getenv(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}

func LoadConfig() Config {
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        log.Fatal("DATABASE_URL is required")
    }
    jwt := getenv("JWT_SECRET", "dev-secret-change-me")
    emailAPIKey := getenv("EMAIL_API_KEY", "")
    emailFromAddress := getenv("EMAIL_FROM_ADDRESS", "noreply@example.com")

    return Config{
        DatabaseDSN: dsn,
        JWTSecret:   jwt,
        EmailAPIKey: emailAPIKey,
        EmailFromAddress: emailFromAddress,
    }
}
