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
			OpenTime  int64       `json:"t"`
			Open      json.Number `json:"o"`
			High      json.Number `json:"h"`
			Low       json.Number `json:"l"`
			Close     json.Number `json:"c"`
			Volume    json.Number `json:"v"`
			CloseTime int64       `json:"T"`
		} `json:"k"`
	} `json:"data"`
}

// HandleKline processes a single raw kline message:
// - updates sliding window
// - asynchronously computes all indicators
// - evaluates alerts and publishes if conditions met
func HandleKline(calcSvc *calculator.Calculator, ctx context.Context, raw []byte) {
	// 1) Ham event’i parse et
	log.Printf("[processor] raw kline event: %s", string(raw))
	var evt RawEvent
	if err := json.Unmarshal(raw, &evt); err != nil {
		log.Printf("processor: invalid raw event: %v", err)
		return
	}

	sym := evt.Data.Symbol
	interval := evt.Data.K.Interval
	key := fmt.Sprintf("%s:%s", sym, interval)

	// Retrieve job
	calcSvc.Mutex().Lock()
	job, ok := calcSvc.Jobs()[key]
	calcSvc.Mutex().Unlock()
	if !ok {
		log.Printf("[processor] no active job for key %s, skipping", key)
		return
	}

	// Current window for this symbol
	window := job.Windows[sym]

	// Build new Kline using raw timestamps
	newK := calculator.Kline{
		OpenTime:  time.UnixMilli(evt.Data.K.OpenTime),
		Open:      calculator.ParseFloat(evt.Data.K.Open),
		High:      calculator.ParseFloat(evt.Data.K.High),
		Low:       calculator.ParseFloat(evt.Data.K.Low),
		Close:     calculator.ParseFloat(evt.Data.K.Close),
		Volume:    calculator.ParseFloat(evt.Data.K.Volume),
		CloseTime: time.UnixMilli(evt.Data.K.CloseTime),
		IsClosed:  evt.Data.K.IsClosed,
	}

	// Update sliding window
	calcSvc.Mutex().Lock()
	if newK.IsClosed {
		window = append(window[1:], newK)
		log.Printf("[processor] appended closed kline for %s, window size now %d", sym, len(job.Windows[sym]))
	} else {
		window[len(window)-1] = newK
		log.Printf("[processor] updated open kline for %s", sym)
	}
	job.Windows[sym] = window
	calcSvc.Mutex().Unlock()

	// Asynchronous indicator computations
	var wg sync.WaitGroup
	resultsCh := make(chan bool, len(job.Indicators))
	messagesCh := make(chan string, len(job.Indicators))

	for _, cfg := range job.Indicators {
		wg.Add(1)
		go func(cfg calculator.IndicatorConfig) {
			defer wg.Done()
			// Compute indicator value
			log.Printf("[processor] computing %s for %s:%s", cfg.Name, sym, interval)
			val := calculator.CalculateIndicator(window, cfg.Name, sym, interval, cfg.Params)
			log.Printf("[processor] %s result for %s:%s = %.4f", cfg.Name, sym, interval, val)
			// Previous value
			prev := calcSvc.GetPrevious(sym, cfg.Name)
			// Evaluate alert condition
			met, err := calculator.EvaluateAlert(val, cfg.Threshold, cfg.Operator, prev)
			if err != nil {
				log.Printf("processor: EvaluateAlert error: %v", err)
				resultsCh <- false
				messagesCh <- fmt.Sprintf("%s error: %v", cfg.Name, err)
				return
			}
			// Log if individual indicator condition met
			if met {
				log.Printf("processor: Indicator '%s' met condition for %s:%s (value=%.4f, threshold=%.4f, operator=%s)", cfg.Name, sym, interval, val, cfg.Threshold, cfg.Operator)
			}
			// Send details
			messagesCh <- fmt.Sprintf("%s => current: %.4f, prev: %.4f, op: %s, thr: %.4f, met: %v",
				cfg.Name, val, prev, cfg.Operator, cfg.Threshold, met)
			resultsCh <- met
			// Store for next iteration
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
	if allMet {
		log.Printf("processor: All indicators met for %s:%s, publishing alert", sym, interval)
		details := ""
		for msg := range messagesCh {
			details += msg + ""
		}

		alertPayload := map[string]interface{}{ // use map[string]interface{} to encode JSON
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
