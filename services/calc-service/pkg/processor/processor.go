// services/calc-service/pkg/processor/processor.go
package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ae144de/sonarbot-service-infra2/services/calc-service/pkg/calculator"
	kafka "github.com/segmentio/kafka-go"
)

// RawEvent represents the kline.raw message structure
type RawEvent struct {
	Data struct {
		Symbol string `json:"s"`
		K      struct {
			Interval  string      `json:"i"`
			IsClosed  bool        `json:"x"`
			Open      json.Number `json:"o"`
			High      json.Number `json:"h"`
			Low       json.Number `json:"l"`
			Close     json.Number `json:"c"`
			Volume    json.Number `json:"v"`
			CloseTime int64       `json:"T"`
			OpenTime  int64       `json:"t"`
		} `json:"k"`
	} `json:"data"`
}

// HandleKline processes a single raw kline message:
// - updates sliding window
// - asynchronously computes all indicators
// - evaluates alerts and publishes if conditions met
func HandleKline(calcSvc *calculator.Calculator, ctx context.Context, raw []byte) {
	var evt RawEvent
	if err := json.Unmarshal(raw, &evt); err != nil {
		log.Printf("processor: invalid raw event: %v", err)
		return
	}

	sym := evt.Data.Symbol
	interval := evt.Data.K.Interval
	key := fmt.Sprintf("%s:%s", sym, interval)

	// Lock, get job and window, then unlock
	calcSvc.Mutex().Lock()
	job, ok := calcSvc.Jobs()[key]
	if !ok {
		calcSvc.Mutex().Unlock()
		return
	}
	window := job.Windows[sym]
	// Build new Kline
	open, _ := evt.Data.K.Open.Float64()
	high, _ := evt.Data.K.High.Float64()
	low, _ := evt.Data.K.Low.Float64()
	closeVal, _ := evt.Data.K.Close.Float64()
	volume, _ := evt.Data.K.Volume.Float64()
	newK := calculator.Kline{
		OpenTime:  time.UnixMilli(evt.Data.K.CloseTime).Add(-calculator.DurationFromInterval(interval)),
		Open:      open,
		High:      high,
		Low:       low,
		Close:     closeVal,
		Volume:    volume,
		CloseTime: time.UnixMilli(evt.Data.K.CloseTime),
		IsClosed:  evt.Data.K.IsClosed,
	}
	// Update sliding window
	if newK.IsClosed {
		window = append(window[1:], newK)
	} else {
		window[len(window)-1] = newK
	}
	job.Windows[sym] = window
	calcSvc.Mutex().Unlock()

	// Asynchronous indicator computation
	var wg sync.WaitGroup
	resultsCh := make(chan bool, len(job.Indicators))
	messagesCh := make(chan string, len(job.Indicators))

	for _, cfg := range job.Indicators {
		wg.Add(1)
		go func(cfg calculator.IndicatorConfig) {
			defer wg.Done()
			// Compute indicator value
			val := calculator.CalculateIndicator(window, cfg.Name, sym, interval, cfg.Params)
			// Get previous value
			prev := calcSvc.GetPrevious(sym, cfg.Name)
			// Evaluate alert condition
			met, err := calculator.EvaluateAlert(val, cfg.Threshold, cfg.Operator, prev)
			if err != nil {
				log.Printf("processor: EvaluateAlert error: %v", err)
				resultsCh <- false
				messagesCh <- fmt.Sprintf("%s error: %v", cfg.Name, err)
				return
			}
			// Build message detail
			messagesCh <- fmt.Sprintf("%s => current: %.4f, prev: %.4f, op: %s, thr: %.4f, met: %v",
				cfg.Name, val, prev, cfg.Operator, cfg.Threshold, met)
			resultsCh <- met
			// Store previous
			calcSvc.SetPrevious(sym, cfg.Name, val)
		}(cfg)
	}
	wg.Wait()
	close(resultsCh)
	close(messagesCh)

	// Aggregate results
	allMet := true
	for met := range resultsCh {
		if !met {
			allMet = false
			break
		}
	}
	// If any met, publish aggregated alert
	if allMet {
		details := ""
		for msg := range messagesCh {
			details += msg + "\n"
		}
		alertPayload := map[string]interface{}{
			"symbol":     sym,
			"interval":   interval,
			"indicators": details,
			"timestamp":  time.Now().Unix(),
		}
		b, _ := json.Marshal(alertPayload)
		if err := calcSvc.Writer().WriteMessages(ctx, kafka.Message{Value: b}); err != nil {
			log.Printf("processor: alert publish error: %v", err)
		}
	}
}
