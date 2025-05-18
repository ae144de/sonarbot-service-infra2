package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/segmentio/kafka-go"
)

// WSKlineResponse represents the websocket kline response
type WSKlineResponse struct {
	Stream string `json:"stream"`
	Data   struct {
		Symbol string `json:"s"`
	} `json:"data"`
}

// SubscribeAndPublish opens WS for given symbols+interval and writes to Kafka.
func SubscribeAndPublish(ctx context.Context, symbols []string, interval string, writer *kafka.Writer) error {
	// Build stream param: btcusdt@kline_1m/...
	streams := make([]string, len(symbols))
	for i, sym := range symbols {
		streams[i] = fmt.Sprintf("%s@kline_%s", strings.ToLower(sym), interval)
	}
	url := fmt.Sprintf("wss://fstream.binance.com/stream?streams=%s", strings.Join(streams, "/"))

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	// for {
	// 	select {
	// 	case <-ctx.Done():
	// 		return nil
	// 	default:
	// 		_, msg, err := conn.ReadMessage()
	// 		if err != nil {
	// 			log.Printf("WS read error: %v", err)
	// 			continue
	// 		}
	// 		// Directly publish raw kline event
	// 		if err := writer.WriteMessages(ctx, kafka.Message{Value: msg}); err != nil {
	// 			log.Printf("Kafka publish error: %v", err)
	// 		}
	// 	}
	// }

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("WS read error: %v", err)
				continue
			}

			// Parse the raw message
			var klineResp WSKlineResponse
			if err := json.Unmarshal(msg, &klineResp); err != nil {
				log.Printf("JSON parse error: %v", err)
				continue
			}

			// Log the kline fetch
			log.Printf("Kline fetched for %s in %s interval.", klineResp.Data.Symbol, interval)

			// Publish raw message to Kafka
			if err := writer.WriteMessages(ctx, kafka.Message{Value: msg}); err != nil {
				log.Printf("Kafka publish error: %v", err)
			}
		}
	}
}
