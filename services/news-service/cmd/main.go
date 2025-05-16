package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ae144de/sonarbot-service-infra2/services/news-service/pkg/news"
	kafka "github.com/segmentio/kafka-go"
)

func main() {
	apiKey := os.Getenv("NEWS_API_KEY")
	fetchInterval := time.Duration(60) * time.Second
	kafkaAddr := os.Getenv("KAFKA_ADDR")
	topic := os.Getenv("KAFKA_NEWS_TOPIC")

	client := news.NewClient(apiKey)
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{kafkaAddr},
		Topic:   topic,
	})
	defer writer.Close()

	ticker := time.NewTicker(fetchInterval)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Println("News Service started")
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			articles, err := client.Fetch(ctx)
			if err != nil {
				log.Printf("Fetch error: %v", err)
				continue
			}
			for _, art := range articles {
				data, _ := json.Marshal(art)
				writer.WriteMessages(ctx, kafka.Message{Value: data})
			}
		}
	}
}
