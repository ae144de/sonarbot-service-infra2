package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ae144de/sonarbot-service-infra2/services/kline-fetcher/pkg/fetcher"
	kafka "github.com/segmentio/kafka-go"
)

func main() {
	// 1) Load config from env
	interval := os.Getenv("INTERVAL")
	if interval == "" {
		interval = "1m"
	}
	groupIdx := fetcher.EnvAsInt("SYMBOL_GROUP", 1)
	totalGroups := fetcher.EnvAsInt("TOTAL_GROUPS", 3)

	kafkaAddr := os.Getenv("KAFKA_ADDR")
	topic := os.Getenv("KAFKA_TOPIC")

	// 2) Kafka writer: async ve ACK beklemeden
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      []string{kafkaAddr},
		Topic:        topic,
		Async:        true,                   // hemen döner
		RequiredAcks: int(kafka.RequireNone), // ACK beklemez
	})
	defer writer.Close()

	// 3) Buffered channel: WebSocket’ten gelen raw mesajları buraya it
	msgCh := make(chan kafka.Message, 10_000)

	// 4) Tek goroutine: kanaldan okuyup Kafka’ya yazar
	go func() {
		for msg := range msgCh {
			if err := writer.WriteMessages(context.Background(), msg); err != nil {
				log.Printf("[kafka-producer] write error: %v", err)
			}
		}
	}()

	// 5) Sembolleri çek ve gruplandır
	symbols, err := fetcher.GetAllFuturesSymbols()
	if err != nil {
		log.Fatalf("Failed to fetch symbols: %v", err)
	}
	mySymbols := fetcher.GetSymbolsForGroup(symbols, groupIdx, totalGroups)
	log.Printf("[%s group %d] subscribing to %d symbols", interval, groupIdx, len(mySymbols))

	// 6) Shutdown context
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// 7) WS subscription & publish loop
	if err := fetcher.SubscribeAndPublish(ctx, mySymbols, interval, msgCh); err != nil {
		log.Fatalf("Subscription error: %v", err)
	}
}
