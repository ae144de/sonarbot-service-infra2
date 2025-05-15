package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds API Gateway configuration.
type Config struct {
	Port         string
	KafkaBroker  string
	KafkaTopic   string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// LoadConfig reads environment variables into Config.
func LoadConfig() Config {
	return Config{
		Port:         getEnv("API_PORT", "8080"),
		KafkaBroker:  getEnv("KAFKA_ADDR", "localhost:9092"),
		KafkaTopic:   getEnv("KAFKA_TOPIC", "kline.raw"),
		ReadTimeout:  time.Second * time.Duration(getEnvAsInt("READ_TIMEOUT", 5)),
		WriteTimeout: time.Second * time.Duration(getEnvAsInt("WRITE_TIMEOUT", 10)),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvAsInt(name string, defaultVal int) int {
	if v := os.Getenv(name); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}
