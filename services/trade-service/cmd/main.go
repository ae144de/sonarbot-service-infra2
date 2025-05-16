package main

// TradeCommand represents an instruction to execute a trade.
type TradeCommand struct {
	Symbol   string  `json:"symbol"`
	Side     string  `json:"side"`
	Quantity float64 `json:"quantity"`
}

func main() {
	// kafkaAddr := os.Getenv("KAFKA_ADDR")
	// topic := os.Getenv("KAFKA_TRADE_TOPIC")

	// reader := kafka.NewReader(kafka.ReaderConfig{
	// 	Brokers: []string{kafkaAddr},
	// 	GroupID: "trade-group",
	// 	Topic:   topic,
	// })
	// defer reader.Close()

	// client := binance.NewClient()

	// ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	// defer cancel()

	// log.Println("Trade Service started, listening for commands")
	// for {
	// 	m, err := reader.FetchMessage(ctx)
	// 	if err != nil {
	// 		break
	// 	}
	// 	var cmd TradeCommand
	// 	if err := json.Unmarshal(m.Value, &cmd); err != nil {
	// 		log.Printf("Unmarshal error: %v", err)
	// 		reader.CommitMessages(ctx, m)
	// 		continue
	// 	}
	// 	// resp, err := client.PlaceOrder(ctx, cmd.Symbol, cmd.Side, cmd.Quantity)
	// 	// if err != nil {
	// 	// 	log.Printf("Order error: %v", err)
	// 	// } else {
	// 	// 	log.Printf("Order placed: %v", resp)
	// 	// }
	// 	reader.CommitMessages(ctx, m)
	// }
}
