package notifier

import (
	"os"
)

// Config holds notifier settings.
type Config struct {
	TelegramToken string
	ChatID        string
	EmailSender   string
	EmailAPIKey   string
}

// LoadConfig reads notifier configs from env.
func LoadConfig() Config {
	return Config{
		TelegramToken: os.Getenv("TELEGRAM_TOKEN"),
		ChatID:        os.Getenv("TELEGRAM_CHAT_ID"),
		EmailSender:   os.Getenv("EMAIL_SENDER"),
		EmailAPIKey:   os.Getenv("EMAIL_API_KEY"),
	}
}
