package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Side represents buy or sell
type Side string

const (
	SideBuy  Side = "buy"
	SideSell Side = "sell"
)

// PositionEffect describes what happened to a trader's position
type PositionEffect string

const (
	EffectOpen        PositionEffect = "open"        // New position opened
	EffectClose       PositionEffect = "close"       // Position closed voluntarily
	EffectLiquidation PositionEffect = "liquidation" // Forced closure
)

// OrderType represents the type of order
type OrderType string

const (
	OrderTypeLimit  OrderType = "limit"
	OrderTypeMarket OrderType = "market"
)

// OrderStatus represents the current state of an order
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPartial   OrderStatus = "partial"
	OrderStatusFilled    OrderStatus = "filled"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// TraderType identifies the kind of participant
type TraderType string

const (
	TraderTypeHuman       TraderType = "human"
	TraderTypeBot         TraderType = "bot"
	TraderTypeMarketMaker TraderType = "market_maker"
)

// Trader represents a market participant
type Trader struct {
	ID          uuid.UUID   `json:"id"`
	Username    string      `json:"username"`
	Type        TraderType  `json:"type"`
	CreatedAt   time.Time   `json:"created_at"`
	TotalPnL    decimal.Decimal `json:"total_pnl"`
	TradeCount  int64       `json:"trade_count"`
}

// Instrument represents a tradeable asset
type Instrument struct {
	Symbol        string          `json:"symbol"`
	BaseCurrency  string          `json:"base_currency"`
	QuoteCurrency string          `json:"quote_currency"`
	TickSize      decimal.Decimal `json:"tick_size"`
	LotSize       decimal.Decimal `json:"lot_size"`
	MaxLeverage   int             `json:"max_leverage"`
}

// Order represents a trading order
type Order struct {
	ID           uuid.UUID       `json:"id"`
	TraderID     uuid.UUID       `json:"trader_id"`
	Instrument   string          `json:"instrument"`
	Side         Side            `json:"side"`
	Type         OrderType       `json:"type"`
	Price        decimal.Decimal `json:"price"`         // Limit price (zero for market)
	Size         decimal.Decimal `json:"size"`          // Original size
	FilledSize   decimal.Decimal `json:"filled_size"`   // How much has been filled
	Status       OrderStatus     `json:"status"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// RemainingSize returns unfilled quantity
func (o *Order) RemainingSize() decimal.Decimal {
	return o.Size.Sub(o.FilledSize)
}

// Trade represents an executed trade - the core of transparency
type Trade struct {
	ID                   uuid.UUID       `json:"id"`
	Instrument           string          `json:"instrument"`
	Price                decimal.Decimal `json:"price"`
	Size                 decimal.Decimal `json:"size"`
	Timestamp            time.Time       `json:"timestamp"`

	// TRANSPARENCY: Both sides are always visible
	BuyerID              uuid.UUID       `json:"buyer_id"`
	SellerID             uuid.UUID       `json:"seller_id"`
	BuyerOrderID         uuid.UUID       `json:"buyer_order_id"`
	SellerOrderID        uuid.UUID       `json:"seller_order_id"`

	// What happened to each trader's position
	BuyerEffect          PositionEffect  `json:"buyer_effect"`
	SellerEffect         PositionEffect  `json:"seller_effect"`

	// New position sizes after this trade
	BuyerNewPosition     decimal.Decimal `json:"buyer_new_position"`
	SellerNewPosition    decimal.Decimal `json:"seller_new_position"`

	// Aggressor side (who took liquidity)
	AggressorSide        Side            `json:"aggressor_side"`
}

// Position represents a trader's current position in an instrument
type Position struct {
	TraderID         uuid.UUID       `json:"trader_id"`
	Instrument       string          `json:"instrument"`
	Size             decimal.Decimal `json:"size"`              // Positive = long, Negative = short
	EntryPrice       decimal.Decimal `json:"entry_price"`       // Average entry price
	UnrealizedPnL    decimal.Decimal `json:"unrealized_pnl"`
	RealizedPnL      decimal.Decimal `json:"realized_pnl"`
	LiquidationPrice decimal.Decimal `json:"liquidation_price"`
	Margin           decimal.Decimal `json:"margin"`
	Leverage         int             `json:"leverage"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// IsLong returns true if position is long
func (p *Position) IsLong() bool {
	return p.Size.IsPositive()
}

// IsShort returns true if position is short
func (p *Position) IsShort() bool {
	return p.Size.IsNegative()
}

// OpenInterestBreakdown provides the transparent OI data
type OpenInterestBreakdown struct {
	Instrument        string    `json:"instrument"`
	Timestamp         time.Time `json:"timestamp"`
	TotalOI           int64     `json:"total_oi"`
	LongPositions     int64     `json:"long_positions"`
	ShortPositions    int64     `json:"short_positions"`

	// Period stats (e.g., last hour, last 24h)
	NewLongsOpened    int64     `json:"new_longs_opened"`
	NewShortsOpened   int64     `json:"new_shorts_opened"`
	LongsClosed       int64     `json:"longs_closed"`
	ShortsClosed      int64     `json:"shorts_closed"`
	LongsLiquidated   int64     `json:"longs_liquidated"`
	ShortsLiquidated  int64     `json:"shorts_liquidated"`
}

// OrderBookLevel represents a price level in the book
type OrderBookLevel struct {
	Price      decimal.Decimal `json:"price"`
	Size       decimal.Decimal `json:"size"`
	OrderCount int             `json:"order_count"`
}

// OrderBook represents the full order book state
type OrderBook struct {
	Instrument string           `json:"instrument"`
	Bids       []OrderBookLevel `json:"bids"` // Sorted high to low
	Asks       []OrderBookLevel `json:"asks"` // Sorted low to high
	Timestamp  time.Time        `json:"timestamp"`
}
