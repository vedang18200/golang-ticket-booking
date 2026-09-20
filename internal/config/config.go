package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port        string
	DatabaseURL string
	RedisAddr   string
	RabbitURL   string
	JWTSecret   string
	SeatHoldTTL time.Duration
	BcryptCost  int
}

// Load reads env vars; the defaults match docker-compose.yml so `make api` just works.
func Load() Config {
	return Config{
		Port:        env("PORT", "8080"),
		DatabaseURL: env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/booking?sslmode=disable"),
		RedisAddr:   env("REDIS_ADDR", "localhost:6379"),
		RabbitURL:   env("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		JWTSecret:   env("JWT_SECRET", "dev-secret-change-me"),
		SeatHoldTTL: envDuration("SEAT_HOLD_TTL", 5*time.Minute),
		BcryptCost:  envInt("BCRYPT_COST", 10), // use 4 for load tests
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
