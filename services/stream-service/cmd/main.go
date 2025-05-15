package main

import (
	"context"

	// "github.com/your-org/platform-monorepo/services/pkg/log"
	"github.com/ae144de/sonarbot-service-infra2/services/stream-service/pkg/kafka"
)

// Streamer handles consuming kline data and publishing to Kafka.
type Streamer struct {
	producer kafka.Producer
}

// NewStreamer creates a new Streamer instance.
func NewStreamer(producer kafka.Producer) *Streamer {
	return &Streamer{producer: producer}
}

// Run starts the streaming process.
func (s *Streamer) Run(ctx context.Context) {
	// TODO: implement WebSocket connection, message handling,
	// update CalculationMap/Slice, publish raw kline to Kafka.
	<-ctx.Done()
}

func main() {
	// logger := log.NewLogger()
	// cfg := kafka.LoadConfig()

	// ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	// defer cancel()

	// producer, err := kafka.NewProducer(cfg)
	// if err != nil {
	// 	logger.Fatal().Err(err).Msg("Kafka producer oluşturulamadı")
	// }

	// stream := NewStreamer(producer)
	// stream.Run(ctx)
}
