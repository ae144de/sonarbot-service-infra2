package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ae144de/sonarbot-service-infra2/services/notify-service/pkg/notifier"
	kafka "github.com/segmentio/kafka-go"
)

func main() {
	// Load config
	cfg := notifier.LoadConfig()
	notifierClient := notifier.New(cfg)

	// Kafka consumer for alerts
	kafkaAddr := os.Getenv("KAFKA_ADDR")
	topic := os.Getenv("KAFKA_ALERT_TOPIC")
	groupID := os.Getenv("KAFKA_GROUP_ID")

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaAddr},
		GroupID: groupID,
		Topic:   topic,
	})
	defer reader.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Println("Notify Service started, consuming from topic:", topic)
	for {
		m, err := reader.FetchMessage(ctx)
		if err != nil {
			break
		}
		// TODO: mesajı parse et, gönderilecek içeriği hazırla
		text := string(m.Value)
		if err := notifierClient.SendTelegram(text); err != nil {
			log.Printf("Telegram send error: %v", err)
		}
		reader.CommitMessages(ctx, m)
	}
}
