package calculator

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	binance "github.com/adshao/go-binance/v2/futures"
	talib "github.com/markcheno/go-talib"
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
	mu         sync.Mutex
	jobs       map[string]*Job
	prevValues map[string]map[string]float64
	client     *binance.Client
	writer     *kafka.Writer
}

// NewCalculator returns a Calculator that will publish alerts.
func NewCalculator(writer *kafka.Writer) *Calculator {
	return &Calculator{
		jobs:       make(map[string]*Job),
		prevValues: make(map[string]map[string]float64),
		client:     binance.NewClient("", ""),
		writer:     writer,
	}
}

// Writer returns the Kafka writer for alerts.
func (c *Calculator) Writer() *kafka.Writer { return c.writer }

// Mutex returns pointer to internal mutex for safe access.
func (c *Calculator) Mutex() *sync.Mutex { return &c.mu }

// Jobs returns the map of active jobs.
func (c *Calculator) Jobs() map[string]*Job { return c.jobs }

// GetPrevious retrieves the last computed value for a symbol and indicator.
func (c *Calculator) GetPrevious(symbol, name string) float64 {
	if m, ok := c.prevValues[symbol]; ok {
		return m[name]
	}
	return 0
}

// SetPrevious stores the computed value for next comparison.
func (c *Calculator) SetPrevious(symbol, name string, value float64) {
	if _, ok := c.prevValues[symbol]; !ok {
		c.prevValues[symbol] = make(map[string]float64)
	}
	c.prevValues[symbol][name] = value
}

// CalculateIndicator computes the specified indicator on the sliding window.
func CalculateIndicator(window []Kline, name, symbol, interval string, params map[string]interface{}) float64 {
	data := make([]float64, len(window))
	for i, k := range window {
		data[i] = k.Close
	}
	switch strings.ToUpper(name) {
	case "RSI":
		if p, ok := params["period"].(float64); ok {
			vals := talib.Rsi(data, int(p))
			if len(vals) > 0 {
				return vals[len(vals)-1]
			}
		}
	case "EMA":
		if p, ok := params["period"].(float64); ok {
			vals := talib.Ema(data, int(p))
			if len(vals) > 0 {
				return vals[len(vals)-1]
			}
		}
		// Extend with additional indicators as needed.
	}
	return 0
}

// EvaluateAlert applies the operator to current and previous values.
func EvaluateAlert(val, threshold float64, operator string, prev float64) (bool, error) {
	op := strings.ToUpper(operator)
	switch op {
	case "GREATER THAN":
		return val > threshold, nil
	case "LESS THAN":
		return val < threshold, nil
	case "CROSSING":
		return (prev < threshold && val >= threshold) || (prev > threshold && val <= threshold), nil
	case "CROSSING UP":
		return prev < threshold && val >= threshold, nil
	case "CROSSING DOWN":
		return prev > threshold && val <= threshold, nil
	default:
		return false, errors.New("unknown operator: " + operator)
	}
}

// ParseFloat safely converts json.Number to float64.
func ParseFloat(n json.Number) float64 {
	f, _ := n.Float64()
	return f
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

// ... rest of the code ...

func atof(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
