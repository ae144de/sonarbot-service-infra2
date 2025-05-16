package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ae144de/sonarbot-service-infra2/services/calc-service/pkg/redis"
	kafka "github.com/segmentio/kafka-go"
)

func main() {
	// Config
	kafkaAddr := os.Getenv("KAFKA_ADDR")
	topic := os.Getenv("KAFKA_TOPIC")
	groupID := os.Getenv("KAFKA_GROUP_ID")

	// Redis init
	if err := redis.Init(); err != nil {
		log.Fatalf("Redis init error: %v", err)
	}

	// Kafka consumer setup
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaAddr},
		GroupID: groupID,
		Topic:   topic,
	})
	defer reader.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Println("Calc Service started, consuming from topic:", topic)
	for {
		m, err := reader.FetchMessage(ctx)
		if err != nil {
			break // context cancelled or error
		}
		// TODO: Deserialize kline, update Redis list, calculate indicators, publish results

		// Commit offset
		if err := reader.CommitMessages(ctx, m); err != nil {
			log.Printf("Commit error: %v", err)
		}
	}
}
