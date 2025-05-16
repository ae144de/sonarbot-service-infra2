package news

import (
	"context"
	"encoding/json"
	"net/http"
)

// Article represents a news item.
type Article struct {
	Title     string `json:"title"`
	Body      string `json:"body"`
	Symbol    string `json:"symbol"`
	Timestamp int64  `json:"timestamp"`
}

// Client fetches news from external API.
type Client struct {
	APIKey  string
	BaseURL string
}

// NewClient creates a new news client.
func NewClient(apiKey string) *Client {
	return &Client{APIKey: apiKey, BaseURL: "https://newsapi.example.com"}
}

// Fetch retrieves latest news articles.
func (c *Client) Fetch(ctx context.Context) ([]Article, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/v1/articles?key="+c.APIKey, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var articles []Article
	if err := json.NewDecoder(resp.Body).Decode(&articles); err != nil {
		return nil, err
	}
	return articles, nil
}
