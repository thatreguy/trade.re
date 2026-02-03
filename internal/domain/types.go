package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Instrument name - single virtual index
const RIndexSymbol = "R.index"

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

// LeverageTier categorizes leverage usage
type LeverageTier string

const (
	LeverageTierConservative LeverageTier = "conservative" // 1-10x
	LeverageTierModerate     LeverageTier = "moderate"     // 11-50x
	LeverageTierAggressive   LeverageTier = "aggressive"   // 51-100x
	LeverageTierDegen        LeverageTier = "degen"        // 101-150x
)

// GetLeverageTier returns the tier for a given leverage
func GetLeverageTier(leverage int) LeverageTier {
	switch {
	case leverage <= 10:
		return LeverageTierConservative
	case leverage <= 50:
		return LeverageTierModerate
	case leverage <= 100:
		return LeverageTierAggressive
	default:
		return LeverageTierDegen
	}
}

// Trader represents a market participant
type Trader struct {
	ID              uuid.UUID       `json:"id"`
	Username        string          `json:"username"`
	Type            TraderType      `json:"type"`
	CreatedAt       time.Time       `json:"created_at"`
	Balance         decimal.Decimal `json:"balance"`          // Available balance
	TotalPnL        decimal.Decimal `json:"total_pnl"`        // Cumulative P&L
	TradeCount      int64           `json:"trade_count"`
	MaxLeverageUsed int             `json:"max_leverage_used"` // Highest leverage ever used (public!)

	// Auth fields (not exposed in JSON)
	PasswordHash    string          `json:"-"`
	APIKeyHash      string          `json:"-"`
}

// Order represents a trading order
type Order struct {
	ID           uuid.UUID       `json:"id"`
	TraderID     uuid.UUID       `json:"trader_id"`
	Instrument   string          `json:"instrument"` // Always "R.index"
	Side         Side            `json:"side"`
	Type         OrderType       `json:"type"`
	Price        decimal.Decimal `json:"price"`         // Limit price (zero for market)
	Size         decimal.Decimal `json:"size"`          // Original size
	FilledSize   decimal.Decimal `json:"filled_size"`   // How much has been filled
	Leverage     int             `json:"leverage"`      // PUBLIC: leverage for this order
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
	Instrument           string          `json:"instrument"` // Always "R.index"
	Price                decimal.Decimal `json:"price"`
	Size                 decimal.Decimal `json:"size"`
	Timestamp            time.Time       `json:"timestamp"`

	// TRANSPARENCY: Both sides are always visible
	BuyerID              uuid.UUID       `json:"buyer_id"`
	SellerID             uuid.UUID       `json:"seller_id"`
	BuyerOrderID         uuid.UUID       `json:"buyer_order_id"`
	SellerOrderID        uuid.UUID       `json:"seller_order_id"`

	// PUBLIC: Leverage used by each side
	BuyerLeverage        int             `json:"buyer_leverage"`
	SellerLeverage       int             `json:"seller_leverage"`

	// What happened to each trader's position
	BuyerEffect          PositionEffect  `json:"buyer_effect"`
	SellerEffect         PositionEffect  `json:"seller_effect"`

	// New position sizes after this trade
	BuyerNewPosition     decimal.Decimal `json:"buyer_new_position"`
	SellerNewPosition    decimal.Decimal `json:"seller_new_position"`

	// Aggressor side (who took liquidity)
	AggressorSide        Side            `json:"aggressor_side"`
}

// Position represents a trader's current position - ALL FIELDS PUBLIC
type Position struct {
	TraderID         uuid.UUID       `json:"trader_id"`
	Instrument       string          `json:"instrument"`        // Always "R.index"
	Size             decimal.Decimal `json:"size"`              // Positive = long, Negative = short
	EntryPrice       decimal.Decimal `json:"entry_price"`       // Average entry price
	Leverage         int             `json:"leverage"`          // PUBLIC: current leverage
	Margin           decimal.Decimal `json:"margin"`            // Margin used
	UnrealizedPnL    decimal.Decimal `json:"unrealized_pnl"`
	RealizedPnL      decimal.Decimal `json:"realized_pnl"`
	LiquidationPrice decimal.Decimal `json:"liquidation_price"` // PUBLIC: where they get liquidated
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

// LeverageTier returns the leverage tier for this position
func (p *Position) LeverageTier() LeverageTier {
	return GetLeverageTier(p.Leverage)
}

// Liquidation records a liquidation event - fully public
type Liquidation struct {
	ID               uuid.UUID       `json:"id"`
	TraderID         uuid.UUID       `json:"trader_id"`
	Instrument       string          `json:"instrument"`
	Side             Side            `json:"side"`              // Long or short that got liquidated
	Size             decimal.Decimal `json:"size"`              // Size liquidated
	EntryPrice       decimal.Decimal `json:"entry_price"`
	LiquidationPrice decimal.Decimal `json:"liquidation_price"`
	MarkPrice        decimal.Decimal `json:"mark_price"`        // Price that triggered liquidation
	Leverage         int             `json:"leverage"`          // PUBLIC: leverage at liquidation
	Loss             decimal.Decimal `json:"loss"`              // Loss from liquidation
	Timestamp        time.Time       `json:"timestamp"`

	// Who took the other side
	CounterpartyID   uuid.UUID       `json:"counterparty_id,omitempty"`
	InsuranceFundHit bool            `json:"insurance_fund_hit"` // Did insurance fund cover?
}

// OpenInterestBreakdown provides the transparent OI data
type OpenInterestBreakdown struct {
	Instrument        string          `json:"instrument"`
	Timestamp         time.Time       `json:"timestamp"`
	TotalOI           decimal.Decimal `json:"total_oi"`
	LongPositions     int64           `json:"long_positions"`
	ShortPositions    int64           `json:"short_positions"`

	// PUBLIC: Average leverage by side
	AvgLongLeverage   decimal.Decimal `json:"avg_long_leverage"`
	AvgShortLeverage  decimal.Decimal `json:"avg_short_leverage"`

	// Period stats
	NewLongsOpened    int64           `json:"new_longs_opened"`
	NewShortsOpened   int64           `json:"new_shorts_opened"`
	LongsClosed       int64           `json:"longs_closed"`
	ShortsClosed      int64           `json:"shorts_closed"`
	LongsLiquidated   int64           `json:"longs_liquidated"`
	ShortsLiquidated  int64           `json:"shorts_liquidated"`
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

// InsuranceFund tracks the insurance fund state
type InsuranceFund struct {
	Balance     decimal.Decimal `json:"balance"`
	TotalIn     decimal.Decimal `json:"total_in"`      // Total added from liquidation profits
	TotalOut    decimal.Decimal `json:"total_out"`     // Total paid out for losses
	UpdatedAt   time.Time       `json:"updated_at"`
}

// MarketStats provides current market statistics
type MarketStats struct {
	Instrument       string          `json:"instrument"`
	LastPrice        decimal.Decimal `json:"last_price"`
	MarkPrice        decimal.Decimal `json:"mark_price"`
	IndexPrice       decimal.Decimal `json:"index_price"` // Same as mark for R.index
	High24h          decimal.Decimal `json:"high_24h"`
	Low24h           decimal.Decimal `json:"low_24h"`
	Volume24h        decimal.Decimal `json:"volume_24h"`
	OpenInterest     decimal.Decimal `json:"open_interest"`
	FundingRate      decimal.Decimal `json:"funding_rate"`
	NextFundingTime  time.Time       `json:"next_funding_time"`
	InsuranceFund    decimal.Decimal `json:"insurance_fund"`
	Timestamp        time.Time       `json:"timestamp"`
}
