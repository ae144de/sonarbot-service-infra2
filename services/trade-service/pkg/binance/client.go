package binance

import (
	"os"

	futures "github.com/adshao/go-binance/v2/futures"
)

// Client wraps Binance Futures API.
type Client struct {
	api *futures.Client
}

// NewClient initializes Binance client with API keys.
func NewClient() *Client {
	return &Client{
		api: futures.NewClient(os.Getenv("BINANCE_API_KEY"), os.Getenv("BINANCE_SECRET_KEY")),
	}
}

// PlaceOrder sends a market order.
// func (c *Client) PlaceOrder(ctx context.Context, symbol, side string, quantity float64) (*futures.CreateOrderResponse, error) {
// 	// return c.api.NewCreateOrderService().
// 	// 	Symbol(symbol).
// 	// 	Side(side).
// 	// 	Type("MARKET").
// 	// 	Quantity(quantity).
// 	// 	Do(ctx)

// }
