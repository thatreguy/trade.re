package liquidation

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/thatreguy/trade.re/internal/config"
	"github.com/thatreguy/trade.re/internal/domain"
)

// PriceProvider gives current market price
type PriceProvider interface {
	GetMarkPrice(instrument string) decimal.Decimal
}

// PositionStore manages positions
type PositionStore interface {
	GetAllPositions(instrument string) []*domain.Position
	GetPosition(traderID uuid.UUID, instrument string) *domain.Position
	ClosePosition(traderID uuid.UUID, instrument string, markPrice decimal.Decimal) error
}

// LiquidationHandler is called when a liquidation occurs
type LiquidationHandler func(liq *domain.Liquidation)

// Engine monitors positions and triggers liquidations
type Engine struct {
	cfg              config.LiquidationConfig
	priceProvider    PriceProvider
	positionStore    PositionStore
	insuranceFund    decimal.Decimal
	insuranceFundMu  sync.RWMutex
	handlers         []LiquidationHandler
	stopCh           chan struct{}
	wg               sync.WaitGroup
}

// NewEngine creates a new liquidation engine
func NewEngine(cfg config.LiquidationConfig, pp PriceProvider, ps PositionStore) *Engine {
	return &Engine{
		cfg:             cfg,
		priceProvider:   pp,
		positionStore:   ps,
		insuranceFund:   cfg.InsuranceFundInitial,
		stopCh:          make(chan struct{}),
	}
}

// OnLiquidation registers a liquidation handler
func (e *Engine) OnLiquidation(handler LiquidationHandler) {
	e.handlers = append(e.handlers, handler)
}

// Start begins the liquidation monitoring loop
func (e *Engine) Start() {
	e.wg.Add(1)
	go e.monitorLoop()
	log.Printf("Liquidation engine started (interval: %dms)", e.cfg.CheckIntervalMs)
}

// Stop halts the liquidation engine
func (e *Engine) Stop() {
	close(e.stopCh)
	e.wg.Wait()
	log.Println("Liquidation engine stopped")
}

// GetInsuranceFund returns current insurance fund balance
func (e *Engine) GetInsuranceFund() decimal.Decimal {
	e.insuranceFundMu.RLock()
	defer e.insuranceFundMu.RUnlock()
	return e.insuranceFund
}

// monitorLoop continuously checks for liquidatable positions
func (e *Engine) monitorLoop() {
	defer e.wg.Done()

	ticker := time.NewTicker(time.Duration(e.cfg.CheckIntervalMs) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-e.stopCh:
			return
		case <-ticker.C:
			e.checkPositions()
		}
	}
}

// checkPositions scans all positions for liquidations
func (e *Engine) checkPositions() {
	markPrice := e.priceProvider.GetMarkPrice(domain.RIndexSymbol)
	if markPrice.IsZero() {
		return // No price available yet
	}

	positions := e.positionStore.GetAllPositions(domain.RIndexSymbol)

	for _, pos := range positions {
		if e.shouldLiquidate(pos, markPrice) {
			e.liquidatePosition(pos, markPrice)
		}
	}
}

// shouldLiquidate determines if a position should be liquidated
func (e *Engine) shouldLiquidate(pos *domain.Position, markPrice decimal.Decimal) bool {
	if pos.Size.IsZero() {
		return false
	}

	if pos.IsLong() {
		// Long position: liquidate if mark price <= liquidation price
		return markPrice.LessThanOrEqual(pos.LiquidationPrice)
	} else {
		// Short position: liquidate if mark price >= liquidation price
		return markPrice.GreaterThanOrEqual(pos.LiquidationPrice)
	}
}

// liquidatePosition executes a liquidation
func (e *Engine) liquidatePosition(pos *domain.Position, markPrice decimal.Decimal) {
	// Calculate loss
	var loss decimal.Decimal
	if pos.IsLong() {
		// Long: loss = (entry - mark) * size
		loss = pos.EntryPrice.Sub(markPrice).Mul(pos.Size)
	} else {
		// Short: loss = (mark - entry) * |size|
		loss = markPrice.Sub(pos.EntryPrice).Mul(pos.Size.Abs())
	}

	// Determine side being liquidated
	var side domain.Side
	if pos.IsLong() {
		side = domain.SideBuy // Long position being liquidated
	} else {
		side = domain.SideSell // Short position being liquidated
	}

	// Create liquidation record
	liq := &domain.Liquidation{
		ID:               uuid.New(),
		TraderID:         pos.TraderID,
		Instrument:       pos.Instrument,
		Side:             side,
		Size:             pos.Size.Abs(),
		EntryPrice:       pos.EntryPrice,
		LiquidationPrice: pos.LiquidationPrice,
		MarkPrice:        markPrice,
		Leverage:         pos.Leverage,
		Loss:             loss,
		Timestamp:        time.Now(),
	}

	// Handle insurance fund
	e.insuranceFundMu.Lock()
	if loss.GreaterThan(pos.Margin) {
		// Loss exceeds margin, insurance fund covers the difference
		shortfall := loss.Sub(pos.Margin)
		if e.insuranceFund.GreaterThanOrEqual(shortfall) {
			e.insuranceFund = e.insuranceFund.Sub(shortfall)
			liq.InsuranceFundHit = true
		} else {
			// Insurance fund depleted - would trigger ADL
			// For now, just use what's available
			e.insuranceFund = decimal.Zero
			liq.InsuranceFundHit = true
			log.Printf("WARNING: Insurance fund depleted during liquidation of %s", pos.TraderID)
		}
	} else {
		// Margin covers the loss, excess goes to insurance fund
		surplus := pos.Margin.Sub(loss)
		e.insuranceFund = e.insuranceFund.Add(surplus)
	}
	e.insuranceFundMu.Unlock()

	// Close the position
	if err := e.positionStore.ClosePosition(pos.TraderID, pos.Instrument, markPrice); err != nil {
		log.Printf("Error closing liquidated position: %v", err)
		return
	}

	// Notify handlers
	for _, handler := range e.handlers {
		handler(liq)
	}

	log.Printf("LIQUIDATION: %s %s %s @ %s (leverage: %dx, loss: %s)",
		pos.TraderID.String()[:8],
		side,
		pos.Size.Abs().String(),
		markPrice.String(),
		pos.Leverage,
		loss.String(),
	)
}

// CalculateLiquidationPrice computes the liquidation price for a position
func CalculateLiquidationPrice(entryPrice decimal.Decimal, leverage int, isLong bool, margins config.MaintenanceMargins) decimal.Decimal {
	maintMargin := margins.GetMarginForLeverage(leverage)
	leverageDecimal := decimal.NewFromInt(int64(leverage))

	// Liquidation distance = entry / leverage * (1 - maintenance margin)
	distance := entryPrice.Div(leverageDecimal).Mul(decimal.NewFromInt(1).Sub(maintMargin))

	if isLong {
		// Long: liquidation price = entry - distance
		return entryPrice.Sub(distance)
	} else {
		// Short: liquidation price = entry + distance
		return entryPrice.Add(distance)
	}
}

// CalculateRequiredMargin computes margin needed for a position
func CalculateRequiredMargin(size, price decimal.Decimal, leverage int) decimal.Decimal {
	notional := size.Abs().Mul(price)
	return notional.Div(decimal.NewFromInt(int64(leverage)))
}

// ValidateLeverage checks if leverage is within allowed range
func ValidateLeverage(leverage, maxLeverage int) bool {
	return leverage >= 1 && leverage <= maxLeverage
}
