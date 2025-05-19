package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"

	"github.com/ae144de/sonarbot-service-infra2/services/api-gateway/pkg/handler"
)

func main() {
	// 1) Config from ENV
	broker := os.Getenv("KAFKA_ADDR")
	topic := os.Getenv("ANALYSIS_REQUEST_TOPIC")
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	if broker == "" || topic == "" {
		log.Fatal("KAFKA_ADDR and ANALYSIS_REQUEST_TOPIC must be set")
	}

	// 2) Kafka Writer
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{broker},
		Topic:   topic,
	})
	defer writer.Close()

	// 3) Gin Engine
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// 4) Register routes
	handler.RegisterRoutes(r, writer, topic)

	// 5) Start server
	addr := ":" + port
	log.Printf("API Gateway listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
