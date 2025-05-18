package fetcher

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

const restURL = "https://fapi.binance.com/fapi/v1/exchangeInfo"

// GetAllFuturesSymbols fetches all USDT futures symbols.
func GetAllFuturesSymbols() ([]string, error) {
	resp, err := http.Get(restURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var info struct {
		Symbols []struct {
			Symbol     string `json:"symbol"`
			QuoteAsset string `json:"quoteAsset"`
			Status     string `json:"status"`
		} `json:"symbols"`
	}
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}

	var symbols []string
	for _, s := range info.Symbols {
		if s.QuoteAsset == "USDT" && s.Status == "TRADING" {
			symbols = append(symbols, s.Symbol)
		}
	}
	return symbols, nil
}

// GetSymbolsForGroup splits symbols into groups of roughly equal size.
func GetSymbolsForGroup(symbols []string, group, total int) []string {
	n := len(symbols)
	size := n / total
	start := (group - 1) * size
	end := start + size
	if group == total {
		end = n
	}
	return symbols[start:end]
}

// EnvAsInt reads an env var or returns default.
func EnvAsInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}
