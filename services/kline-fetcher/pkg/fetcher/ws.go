package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/gorilla/websocket"
	kafka "github.com/segmentio/kafka-go"
)

// WSKlineResponse yalnızca sembolü çıkarmak için
type WSKlineResponse struct {
	Data struct {
		Symbol string `json:"s"`
	} `json:"data"`
}

// SubscribeAndPublish, verilen sembollerin interval'ındaki kline stream'lerini açar
// ve her ham mesajı msgCh kanalına iter.
func SubscribeAndPublish(
	ctx context.Context,
	symbols []string,
	interval string,
	msgCh chan<- kafka.Message,
) error {
	// 1) Stream parametresini oluştur
	streams := make([]string, len(symbols))
	for i, sym := range symbols {
		streams[i] = fmt.Sprintf("%s@kline_%s", strings.ToLower(sym), interval)
	}
	url := fmt.Sprintf("wss://fstream.binance.com/stream?streams=%s", strings.Join(streams, "/"))

	// 2) WebSocket'e bağlan
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	defer conn.Close()
	log.Printf("[kline-fetcher] WS connected: %s", url)

	// 3) Mesajları oku ve kanala iter
	for {
		select {
		case <-ctx.Done():
			log.Printf("[kline-fetcher] context done, exiting")
			return nil
		default:
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("[kline-fetcher] WS read error: %v", err)
				continue
			}

			// İsteğe bağlı: sadece loglamak için sembolü parse et
			var resp WSKlineResponse
			if err := json.Unmarshal(msg, &resp); err == nil {
				log.Printf("[kline-fetcher] fetched kline for %s", resp.Data.Symbol)
			}

			// 4) Ham mesajı Kafka kanalına it
			msgCh <- kafka.Message{Value: msg}
		}
	}
}
