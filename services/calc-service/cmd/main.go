package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ae144de/sonarbot-service-infra2/services/calc-service/pkg/redis"
	"github.com/ae144de/sonarbot-service-infra2/services/calc-service/service"
	kafka "github.com/segmentio/kafka-go"
)

func main() {
	// 1) Redis’i başlat (window state için)
	if err := redis.Init(); err != nil {
		log.Fatalf("Redis init error: %v", err)
	}

	// 2) Kontrol mesajları için Kafka reader (analysis.request)
	ctrlReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{os.Getenv("KAFKA_ADDR")},
		GroupID: "calc-control",
		Topic:   os.Getenv("ANALYSIS_REQUEST_TOPIC"),
	})
	defer ctrlReader.Close()

	// 3) Veri mesajları için Kafka reader (kline.raw)
	dataReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{os.Getenv("KAFKA_ADDR")},
		GroupID: "calc-data",
		Topic:   os.Getenv("KAFKA_TOPIC"),
	})
	defer dataReader.Close()

	// 4) İş mantarı servisini başlat
	calcSvc := service.NewCalculator()

	// 5) Context ve sinyal yakalama
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// 6) Kontrol döngüsü (arama gelen istekleri alıp job tanımlarını oluşturur)
	go func() {
		for {
			m, err := ctrlReader.FetchMessage(ctx)
			if err != nil {
				log.Printf("Control fetch error: %v", err)
				continue
			}
			calcSvc.HandleControl(m.Value)
			ctrlReader.CommitMessages(ctx, m)
		}
	}()

	// 7) Veri döngüsü (kline.raw’dan gelen her kline’ı işleme alır)
	log.Println("Calc-service up, waiting for klines…")
	for {
		m, err := dataReader.FetchMessage(ctx)
		if err != nil {
			log.Printf("Data fetch error: %v", err)
			continue
		}
		calcSvc.HandleKline(m.Value)
		dataReader.CommitMessages(ctx, m)
	}
}
