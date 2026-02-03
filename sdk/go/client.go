// Package tradere provides a Go SDK for the Trade.re trading game API.
package tradere

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
)

const (
	defaultTimeout = 10 * time.Second
	wsPath         = "/ws"
)

// Client is the Trade.re API client
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Trade.re client
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// Side represents buy or sell
type Side string

const (
	SideBuy  Side = "buy"
	SideSell Side = "sell"
)

// OrderType represents order type
type OrderType string

const (
	OrderTypeLimit  OrderType = "limit"
	OrderTypeMarket OrderType = "market"
)

// Trader represents a market participant
type Trader struct {
	ID              string          `json:"id"`
	Username        string          `json:"username"`
	Type            string          `json:"type"`
	Balance         decimal.Decimal `json:"balance"`
	TotalPnL        decimal.Decimal `json:"total_pnl"`
	TradeCount      int64           `json:"trade_count"`
	MaxLeverageUsed int             `json:"max_leverage_used"`
}

// Position represents an open position (PUBLIC!)
type Position struct {
	TraderID         string          `json:"trader_id"`
	Instrument       string          `json:"instrument"`
	Size             decimal.Decimal `json:"size"`
	EntryPrice       decimal.Decimal `json:"entry_price"`
	Leverage         int             `json:"leverage"` // PUBLIC!
	Margin           decimal.Decimal `json:"margin"`
	UnrealizedPnL    decimal.Decimal `json:"unrealized_pnl"`
	RealizedPnL      decimal.Decimal `json:"realized_pnl"`
	LiquidationPrice decimal.Decimal `json:"liquidation_price"`
}

// Order represents a trading order
type Order struct {
	ID         string          `json:"id"`
	TraderID   string          `json:"trader_id"`
	Instrument string          `json:"instrument"`
	Side       Side            `json:"side"`
	Type       OrderType       `json:"type"`
	Price      decimal.Decimal `json:"price"`
	Size       decimal.Decimal `json:"size"`
	FilledSize decimal.Decimal `json:"filled_size"`
	Leverage   int             `json:"leverage"` // PUBLIC!
	Status     string          `json:"status"`
}

// Trade represents an executed trade (TRANSPARENT!)
type Trade struct {
	ID             string          `json:"id"`
	Instrument     string          `json:"instrument"`
	Price          decimal.Decimal `json:"price"`
	Size           decimal.Decimal `json:"size"`
	BuyerID        string          `json:"buyer_id"`
	SellerID       string          `json:"seller_id"`
	BuyerLeverage  int             `json:"buyer_leverage"`  // PUBLIC!
	SellerLeverage int             `json:"seller_leverage"` // PUBLIC!
	BuyerEffect    string          `json:"buyer_effect"`
	SellerEffect   string          `json:"seller_effect"`
	AggressorSide  Side            `json:"aggressor_side"`
	Timestamp      time.Time       `json:"timestamp"`
}

// OrderBookLevel represents a price level
type OrderBookLevel struct {
	Price      decimal.Decimal `json:"price"`
	Size       decimal.Decimal `json:"size"`
	OrderCount int             `json:"order_count"`
}

// OrderBook represents the order book
type OrderBook struct {
	Instrument string           `json:"instrument"`
	Bids       []OrderBookLevel `json:"bids"`
	Asks       []OrderBookLevel `json:"asks"`
}

// OpenInterest represents OI breakdown
type OpenInterest struct {
	Instrument       string          `json:"instrument"`
	TotalOI          decimal.Decimal `json:"total_oi"`
	LongPositions    int64           `json:"long_positions"`
	ShortPositions   int64           `json:"short_positions"`
	AvgLongLeverage  decimal.Decimal `json:"avg_long_leverage"`  // PUBLIC!
	AvgShortLeverage decimal.Decimal `json:"avg_short_leverage"` // PUBLIC!
}

// PlaceOrderRequest is the request to place an order
type PlaceOrderRequest struct {
	Side     Side            `json:"side"`
	Type     OrderType       `json:"type"`
	Price    decimal.Decimal `json:"price,omitempty"`
	Size     decimal.Decimal `json:"size"`
	Leverage int             `json:"leverage"`
}

// PlaceOrderResponse is the response from placing an order
type PlaceOrderResponse struct {
	Order  *Order   `json:"order"`
	Trades []*Trade `json:"trades"`
}

// request makes an HTTP request to the API
func (c *Client) request(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}

	return nil
}

// GetOrderBook retrieves the current order book
func (c *Client) GetOrderBook(ctx context.Context) (*OrderBook, error) {
	var book OrderBook
	if err := c.request(ctx, "GET", "/api/v1/market/orderbook", nil, &book); err != nil {
		return nil, err
	}
	return &book, nil
}

// GetAllPositions retrieves ALL positions - transparency!
func (c *Client) GetAllPositions(ctx context.Context) ([]*Position, error) {
	var positions []*Position
	if err := c.request(ctx, "GET", "/api/v1/market/positions", nil, &positions); err != nil {
		return nil, err
	}
	return positions, nil
}

// GetOpenInterest retrieves OI breakdown with leverage stats
func (c *Client) GetOpenInterest(ctx context.Context) (*OpenInterest, error) {
	var oi OpenInterest
	if err := c.request(ctx, "GET", "/api/v1/market/oi", nil, &oi); err != nil {
		return nil, err
	}
	return &oi, nil
}

// GetRecentTrades retrieves recent trades
func (c *Client) GetRecentTrades(ctx context.Context) ([]*Trade, error) {
	var trades []*Trade
	if err := c.request(ctx, "GET", "/api/v1/market/trades", nil, &trades); err != nil {
		return nil, err
	}
	return trades, nil
}

// GetTraders retrieves all traders (public info)
func (c *Client) GetTraders(ctx context.Context) ([]*Trader, error) {
	var traders []*Trader
	if err := c.request(ctx, "GET", "/api/v1/traders", nil, &traders); err != nil {
		return nil, err
	}
	return traders, nil
}

// GetTrader retrieves a specific trader
func (c *Client) GetTrader(ctx context.Context, traderID string) (*Trader, error) {
	var trader Trader
	if err := c.request(ctx, "GET", "/api/v1/traders/"+traderID, nil, &trader); err != nil {
		return nil, err
	}
	return &trader, nil
}

// GetTraderPositions retrieves a trader's positions (public!)
func (c *Client) GetTraderPositions(ctx context.Context, traderID string) ([]*Position, error) {
	var positions []*Position
	if err := c.request(ctx, "GET", "/api/v1/traders/"+traderID+"/positions", nil, &positions); err != nil {
		return nil, err
	}
	return positions, nil
}

// PlaceOrder submits a new order (requires API key)
func (c *Client) PlaceOrder(ctx context.Context, req PlaceOrderRequest) (*PlaceOrderResponse, error) {
	var resp PlaceOrderResponse
	if err := c.request(ctx, "POST", "/api/v1/orders", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CancelOrder cancels an existing order (requires API key)
func (c *Client) CancelOrder(ctx context.Context, orderID string) error {
	return c.request(ctx, "DELETE", "/api/v1/orders/"+orderID, nil, nil)
}

// WebSocket streaming

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type      string          `json:"type"`
	Channel   string          `json:"channel,omitempty"`
	Data      json.RawMessage `json:"data"`
	Timestamp int64           `json:"timestamp"`
}

// StreamClient handles WebSocket connections
type StreamClient struct {
	conn *websocket.Conn
}

// NewStreamClient creates a WebSocket connection
func (c *Client) NewStreamClient(ctx context.Context) (*StreamClient, error) {
	wsURL := "ws" + c.baseURL[4:] + wsPath // Convert http to ws
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("connecting to WebSocket: %w", err)
	}
	return &StreamClient{conn: conn}, nil
}

// Subscribe subscribes to a channel
func (s *StreamClient) Subscribe(channel string) error {
	msg := map[string]interface{}{
		"type": "subscribe",
		"data": channel,
	}
	return s.conn.WriteJSON(msg)
}

// Read reads the next message
func (s *StreamClient) Read() (*WSMessage, error) {
	var msg WSMessage
	if err := s.conn.ReadJSON(&msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// Close closes the WebSocket connection
func (s *StreamClient) Close() error {
	return s.conn.Close()
}
