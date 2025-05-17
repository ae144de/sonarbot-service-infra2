// package main

// import "log"

//	func main() {
//		log.Println("🔥 API Gateway starting…")
//	}
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ae144de/sonarbot-service-infra2/services/api-gateway/pkg/config"
	"github.com/ae144de/sonarbot-service-infra2/services/api-gateway/pkg/handler"
	"github.com/segmentio/kafka-go"
)

func main() {
	cfg := config.LoadConfig()

	// Kafka producer for analysis requests
	analysisWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{cfg.KafkaBroker},
		Topic:   os.Getenv("ANALYSIS_REQUEST_TOPIC"),
	})
	defer analysisWriter.Close()

	// Kafka producer for test requests
	testWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{cfg.KafkaBroker},
		Topic:   os.Getenv("TEST_REQUEST_TOPIC"),
	})
	defer testWriter.Close()

	// Create router
	router := handler.NewRouter()

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	log.Printf("API Gateway listening on port %s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
