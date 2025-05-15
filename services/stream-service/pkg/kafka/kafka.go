package kafka

import (
	"context"
	"os"

	kafka "github.com/segmentio/kafka-go"
)

// Config holds Kafka connection and topic settings.
type Config struct {
	BrokerAddress string
	Topic         string
}

// LoadConfig reads configuration from environment variables.
func LoadConfig() Config {
	return Config{
		BrokerAddress: os.Getenv("KAFKA_ADDR"),
		Topic:         os.Getenv("KAFKA_TOPIC"),
	}
}

// Producer wraps a kafka.Writer for publishing messages.
type Producer struct {
	Writer *kafka.Writer
}

// NewProducer creates and returns a new Kafka Producer.
func NewProducer(cfg Config) (*Producer, error) {
	writer := &kafka.Writer{
		Addr:  kafka.TCP(cfg.BrokerAddress),
		Topic: cfg.Topic,
	}
	return &Producer{Writer: writer}, nil
}

// Write publishes messages to Kafka.
func (p *Producer) Write(ctx context.Context, msgs ...kafka.Message) error {
	return p.Writer.WriteMessages(ctx, msgs...)
}

// Close closes the Kafka writer.
func (p *Producer) Close() error {
	return p.Writer.Close()
}
