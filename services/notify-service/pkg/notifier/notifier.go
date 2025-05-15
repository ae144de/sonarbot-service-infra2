package notifier

import (
	"fmt"
	"net/http"
)

// Notifier defines methods for sending notifications.
type Notifier struct {
	cfg Config
}

// New creates a new Notifier.
func New(cfg Config) *Notifier {
	return &Notifier{cfg: cfg}
}

// SendTelegram sends a message via Telegram Bot API.
func (n *Notifier) SendTelegram(msg string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s",
		n.cfg.TelegramToken, n.cfg.ChatID, msg)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API status: %s", resp.Status)
	}
	return nil
}
