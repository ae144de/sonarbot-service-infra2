package calculator

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	binance "github.com/adshao/go-binance/v2/futures"
	kafka "github.com/segmentio/kafka-go"
)

// AnalysisRequest models the incoming control payload.
type AnalysisRequest struct {
	WebsocketKlineOptions struct {
		Symbol   string `json:"symbol"`
		Interval string `json:"interval"`
		Exchange string `json:"exchange"`
	} `json:"websocketKlineOptions"`
	Indicators []struct {
		Indicator  string                 `json:"indicator"`
		Parameters map[string]interface{} `json:"parameters"`
		Operator   string                 `json:"operator"`
		Threshold  float64                `json:"threshold"`
	} `json:"indicators"`
}

// Kline is a simplified OHLCV struct.
type Kline struct {
	OpenTime  time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	CloseTime time.Time
	IsClosed  bool
}

// Job holds sliding windows and indicator configs per symbol.
type Job struct {
	Interval   string
	Symbols    []string
	Indicators []IndicatorConfig
	Windows    map[string][]Kline
}

// IndicatorConfig holds what to compute and when to alert.
type IndicatorConfig struct {
	Name      string
	Operator  string
	Threshold float64
	Params    map[string]interface{}
}

// Calculator keeps all active jobs.
type Calculator struct {
	mu     sync.Mutex
	jobs   map[string]*Job
	client *binance.Client
	writer *kafka.Writer
}

// NewCalculator returns a Calculator that will publish alerts.
func NewCalculator(writer *kafka.Writer) *Calculator {
	return &Calculator{
		jobs:   make(map[string]*Job),
		client: binance.NewClient("", ""),
		writer: writer,
	}
}

// HandleControl parses the control message and initializes a Job.
func (c *Calculator) HandleControl(ctx context.Context, raw []byte) {
	var req AnalysisRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		log.Printf("Invalid control payload: %v", err)
		return
	}
	key := req.WebsocketKlineOptions.Symbol + ":" + req.WebsocketKlineOptions.Interval

	// Determine symbols list
	syms := []string{req.WebsocketKlineOptions.Symbol}
	if req.WebsocketKlineOptions.Symbol == "ALL" {
		info, err := c.client.NewExchangeInfoService().Do(ctx)
		if err != nil {
			log.Printf("ExchangeInfo error: %v", err)
			return
		}
		syms = []string{}
		for _, si := range info.Symbols {
			if si.QuoteAsset == "USDT" && si.Status == "TRADING" {
				syms = append(syms, si.Symbol)
			}
		}
	}

	// Fetch historical klines and init windows
	windows := make(map[string][]Kline)
	for _, sym := range syms {
		ks, err := c.client.NewKlinesService().
			Symbol(sym).
			Interval(req.WebsocketKlineOptions.Interval).
			Limit(100).
			Do(ctx)
		if err != nil {
			log.Printf("Hist fetch error for %s: %v", sym, err)
			continue
		}
		arr := make([]Kline, len(ks))
		for i, b := range ks {
			arr[i] = Kline{
				OpenTime:  time.UnixMilli(b.OpenTime),
				Open:      atof(b.Open),
				High:      atof(b.High),
				Low:       atof(b.Low),
				Close:     atof(b.Close),
				Volume:    atof(b.Volume),
				CloseTime: time.UnixMilli(b.CloseTime),
				IsClosed:  b.IsFinal,
			}
		}
		windows[sym] = arr
	}

	// Build indicator configs
	cfgs := make([]IndicatorConfig, len(req.Indicators))
	for i, ind := range req.Indicators {
		cfgs[i] = IndicatorConfig{
			Name:      ind.Indicator,
			Operator:  ind.Operator,
			Threshold: ind.Threshold,
			Params:    ind.Parameters,
		}
	}

	c.mu.Lock()
	c.jobs[key] = &Job{
		Interval:   req.WebsocketKlineOptions.Interval,
		Symbols:    syms,
		Indicators: cfgs,
		Windows:    windows,
	}
	c.mu.Unlock()
}

func atof(n json.Number) float64 {
	f, _ := n.Float64()
	return f
}
