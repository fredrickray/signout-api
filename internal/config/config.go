package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv             string
	HTTPAddr           string
	HTTPReadTimeout    time.Duration
	HTTPWriteTimeout   time.Duration
	HTTPIdleTimeout    time.Duration
	MongoURI           string
	MongoDB            string
	JWTAccessSecret    string
	JWTRefreshSecret   string
	JWTAccessTTL       time.Duration
	JWTRefreshTTL      time.Duration
	CORSAllowedOrigins []string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		HTTPAddr:           getEnv("HTTP_ADDR", ":8080"),
		HTTPReadTimeout:    getDuration("HTTP_READ_TIMEOUT", 10*time.Second),
		HTTPWriteTimeout:   getDuration("HTTP_WRITE_TIMEOUT", 30*time.Second),
		HTTPIdleTimeout:    getDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
		MongoURI:           getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDB:            getEnv("MONGODB_DB", "signout"),
		JWTAccessSecret:    os.Getenv("JWT_ACCESS_SECRET"),
		JWTRefreshSecret:   os.Getenv("JWT_REFRESH_SECRET"),
		JWTAccessTTL:       getDuration("JWT_ACCESS_TTL", 15*time.Minute),
		JWTRefreshTTL:      getDuration("JWT_REFRESH_TTL", 7*24*time.Hour),
		CORSAllowedOrigins: splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")),
	}

	if cfg.MongoURI == "" {
		return Config{}, fmt.Errorf("MONGODB_URI is required")
	}
	if len(cfg.JWTAccessSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_ACCESS_SECRET must be at least 32 characters")
	}
	if len(cfg.JWTRefreshSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_REFRESH_SECRET must be at least 32 characters")
	}

	return cfg, nil
}

func (c Config) IsDev() bool {
	return c.AppEnv == "development" || c.AppEnv == "dev"
}

func getEnv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		if n, err2 := strconv.Atoi(raw); err2 == nil {
			return time.Duration(n) * time.Second
		}
		return fallback
	}
	return d
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
