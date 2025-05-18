package main

import (
	"context"

	"github.com/ae144de/sonarbot-service-infra2/services/services/pkg/log"
	"github.com/ae144de/sonarbot-service-infra2/services/stream-service/pkg/kafka"

	// "os"
	"os/signal"
	"syscall"
)

// Streamer handles consuming kline data and publishing to Kafka.
type Streamer struct {
	producer *kafka.Producer
}

// NewStreamer creates a new Streamer instance.
func NewStreamer(producer *kafka.Producer) *Streamer {
	return &Streamer{producer: producer}
}

// Run starts the streaming process and blocks until context is cancelled.
func (s *Streamer) Run(ctx context.Context) {
	// TODO: implement WebSocket connection, message handling,
	// sliding window update, and publish raw kline to Kafka.
	<-ctx.Done()
}

func main() {
	logger := log.NewLogger()
	cfg := kafka.LoadConfig()

	// Initialize Kafka producer
	producer, err := kafka.NewProducer(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create Kafka producer")
	}
	defer producer.Close()

	// Setup signal handling
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger.Info().Msg("Stream-Service starting…")
	streamer := NewStreamer(producer)
	streamer.Run(ctx)
}
