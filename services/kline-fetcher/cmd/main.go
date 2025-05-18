package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ae144de/sonarbot-service-infra2/services/kline-fetcher/pkg/fetcher"
	"github.com/segmentio/kafka-go"
)

func main() {
	// Load environment
	interval := os.Getenv("INTERVAL")
	if interval == "" {
		interval = "1m"
	}
	groupIdx := fetcher.EnvAsInt("SYMBOL_GROUP", 1)
	totalGroups := fetcher.EnvAsInt("TOTAL_GROUPS", 3)

	// Kafka producer setup
	kafkaAddr := os.Getenv("KAFKA_ADDR")
	topic := os.Getenv("KAFKA_TOPIC")
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{kafkaAddr},
		Topic:   topic,
	})
	defer writer.Close()

	// Fetch and group symbols
	symbols, err := fetcher.GetAllFuturesSymbols()
	if err != nil {
		log.Fatalf("Failed to fetch symbols: %v", err)
	}
	mySymbols := fetcher.GetSymbolsForGroup(symbols, groupIdx, totalGroups)
	log.Printf("[%s group %d] Subscribing to %d symbols", interval, groupIdx, len(mySymbols))

	// Context and signal handling
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Start WS subscription loop
	if err := fetcher.SubscribeAndPublish(ctx, mySymbols, interval, writer); err != nil {
		log.Fatalf("Subscription error: %v", err)
	}
}
