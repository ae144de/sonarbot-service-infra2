package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	kafka "github.com/segmentio/kafka-go"

	"github.com/ae144de/sonarbot-service-infra2/services/calc-service/pkg/calculator"
	"github.com/ae144de/sonarbot-service-infra2/services/calc-service/pkg/processor"
	"github.com/ae144de/sonarbot-service-infra2/services/calc-service/pkg/redis"
)

func main() {

	// A) Sadece shut-down sinyallerini dinleyecek context
	ctxShutdown, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// B) Kafka için hep canlı kalacak background context
	ctxKafka := context.Background()

	// 1) Redis’i başlat (sliding-window için state store)
	if err := redis.Init(); err != nil {
		log.Fatalf("Redis init error: %v", err)
	}

	kafkaAddr := os.Getenv("KAFKA_ADDR")
	rawTopic := os.Getenv("KAFKA_TOPIC")
	ctrlTopic := os.Getenv("ANALYSIS_REQUEST_TOPIC")
	alertTopic := os.Getenv("ALERT_TRIGGER_TOPIC")

	// 2) Control reader (analysis.request)
	ctrlReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{kafkaAddr},
		GroupID:     "calc-control",
		Topic:       ctrlTopic,
		StartOffset: kafka.FirstOffset,
	})
	// defer ctrlReader.Close()

	// 3) Data reader (kline.raw)
	dataReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{kafkaAddr},
		GroupID:     "calc-data",
		Topic:       rawTopic,
		StartOffset: kafka.FirstOffset,
	})
	// defer dataReader.Close()

	// 4) Kafka writer for alerts
	alertWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{kafkaAddr},
		Topic:   alertTopic,
	})
	defer alertWriter.Close()

	// 5) Calculator service
	calcSvc := calculator.NewCalculator(alertWriter)

	// // 6) Context & signal
	// ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	// defer cancel()

	// 7) Control loop: parse incoming analysis.request
	go func() {
		log.Println("▶️ Control loop started, listening for analysis requests…")

		for {
			m, err := ctrlReader.FetchMessage(ctxKafka)
			if err != nil {
				log.Printf("Control fetch error: %v", err)
				continue
			}
			calcSvc.HandleControl(ctxKafka, m.Value)
			ctrlReader.CommitMessages(ctxKafka, m)
		}
	}()

	// 8) Data loop: consume kline.raw and process
	// log.Println("Calc-service up, awaiting kline.raw messages…")
	// for {
	// 	m, err := dataReader.FetchMessage(ctx)
	// 	if err != nil {
	// 		log.Printf("Data fetch error: %v", err)
	// 		continue
	// 	}
	// 	processor.HandleKline(calcSvc, ctx, m.Value)
	// 	dataReader.CommitMessages(ctx, m)
	// }
	// Data loop:
	go func() {
		log.Println("▶️ Data loop started")
		for {
			m, err := dataReader.FetchMessage(ctxKafka)
			if err != nil {
				log.Printf("Data fetch error (will retry): %v", err)
				// time.Sleep(time.Second)
				continue
			}
			processor.HandleKline(calcSvc, ctxKafka, m.Value)
			dataReader.CommitMessages(ctxKafka, m)
		}
	}()

	// Son olarak: shutdown sinyali bekle
	<-ctxShutdown.Done()
	log.Println("Shutting down calc-service…")
	// reader’ları kapatıp exit edebilirsiniz
	ctrlReader.Close()
	dataReader.Close()
}
