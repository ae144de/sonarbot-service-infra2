package config

import (
	"os"
	"time"
)

// Config holds auth service configuration.
type Config struct {
	MongoURI      string
	JWTSigningKey string
	TokenTTL      time.Duration
}

// LoadConfig reads env vars into Config.
func LoadConfig() Config {
	ttl, err := time.ParseDuration(getEnv("TOKEN_TTL", "24h"))
	if err != nil {
		ttl = 24 * time.Hour
	}
	return Config{
		MongoURI:      getEnv("MONGO_URI", "mongodb://localhost:27017"),
		JWTSigningKey: getEnv("JWT_SIGNING_KEY", "supersecret"),
		TokenTTL:      ttl,
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
