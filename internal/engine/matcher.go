package engine

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/thatreguy/trade.re/internal/config"
	"github.com/thatreguy/trade.re/internal/db"
	"github.com/thatreguy/trade.re/internal/domain"
)

// TradeHandler is called when a trade is executed
type TradeHandler func(trade *domain.Trade)

// OrderHandler is called when an order is updated
type OrderHandler func(order *domain.Order)

// LiquidationHandler is called when a liquidation occurs
type LiquidationHandler func(liq *domain.Liquidation)

// MatchingEngine handles order matching for all instruments
type MatchingEngine struct {
	books               map[string]*OrderBook
	positions           map[string]*domain.Position // key: traderID:instrument
	traders             map[uuid.UUID]*domain.Trader
	recentTrades        []*domain.Trade       // Recent trades for history
	liquidations        []*domain.Liquidation // Liquidation history
	mu                  sync.RWMutex
	tradeHandlers       []TradeHandler
	orderHandlers       []OrderHandler
	liquidationHandlers []LiquidationHandler
	db                  *db.SQLiteDB // Optional database for persistence
	liqConfig           *config.LiquidationConfig
}

// NewMatchingEngine creates a new matching engine
func NewMatchingEngine() *MatchingEngine {
	return &MatchingEngine{
		books:        make(map[string]*OrderBook),
		positions:    make(map[string]*domain.Position),
		traders:      make(map[uuid.UUID]*domain.Trader),
		recentTrades: make([]*domain.Trade, 0),
		liquidations: make([]*domain.Liquidation, 0),
	}
}

// SetDatabase sets the SQLite database for persistence
func (me *MatchingEngine) SetDatabase(database *db.SQLiteDB) {
	me.db = database
}

// LoadFromDatabase loads all data from the database
func (me *MatchingEngine) LoadFromDatabase() error {
	if me.db == nil {
		return nil
	}

	me.mu.Lock()
	defer me.mu.Unlock()

	// Load traders
	traders, err := me.db.GetAllTraders()
	if err != nil {
		return fmt.Errorf("loading traders: %w", err)
	}
	for _, t := range traders {
		me.traders[t.ID] = t
	}
	log.Printf("Loaded %d traders from database", len(traders))

	// Load positions for R.index
	positions, err := me.db.GetAllPositions("R.index")
	if err != nil {
		return fmt.Errorf("loading positions: %w", err)
	}
	for _, p := range positions {
		posKey := fmt.Sprintf("%s:%s", p.TraderID, p.Instrument)
		me.positions[posKey] = p
	}
	log.Printf("Loaded %d positions from database", len(positions))

	// Load recent trades
	trades, err := me.db.GetRecentTrades("R.index", 1000)
	if err != nil {
		return fmt.Errorf("loading trades: %w", err)
	}
	me.recentTrades = trades
	log.Printf("Loaded %d trades from database", len(trades))

	// Load recent liquidations
	liquidations, err := me.db.GetRecentLiquidations("R.index", 100)
	if err != nil {
		return fmt.Errorf("loading liquidations: %w", err)
	}
	me.liquidations = liquidations
	log.Printf("Loaded %d liquidations from database", len(liquidations))

	// Load open orders and rebuild order book
	orders, err := me.db.GetOpenOrders("R.index")
	if err != nil {
		return fmt.Errorf("loading orders: %w", err)
	}
	if book, exists := me.books["R.index"]; exists {
		for _, order := range orders {
			book.AddOrder(order)
		}
		log.Printf("Loaded %d open orders from database", len(orders))
	}

	return nil
}

// RegisterInstrument creates an order book for an instrument
func (me *MatchingEngine) RegisterInstrument(instrument string) {
	me.mu.Lock()
	defer me.mu.Unlock()
	if _, exists := me.books[instrument]; !exists {
		me.books[instrument] = NewOrderBook(instrument)
	}
}

// RegisterTrader adds a trader to the system
func (me *MatchingEngine) RegisterTrader(trader *domain.Trader) {
	me.mu.Lock()
	defer me.mu.Unlock()
	me.traders[trader.ID] = trader

	// Persist to database
	if me.db != nil {
		if err := me.db.SaveTrader(trader); err != nil {
			log.Printf("Error saving trader to database: %v", err)
		}
	}
}

// OnTrade registers a trade handler
func (me *MatchingEngine) OnTrade(handler TradeHandler) {
	me.tradeHandlers = append(me.tradeHandlers, handler)
}

// OnOrderUpdate registers an order update handler
func (me *MatchingEngine) OnOrderUpdate(handler OrderHandler) {
	me.orderHandlers = append(me.orderHandlers, handler)
}

// SubmitOrder processes a new order through the matching engine
func (me *MatchingEngine) SubmitOrder(order *domain.Order) ([]*domain.Trade, error) {
	me.mu.Lock()
	defer me.mu.Unlock()

	book, exists := me.books[order.Instrument]
	if !exists {
		return nil, fmt.Errorf("unknown instrument: %s", order.Instrument)
	}

	if _, exists := me.traders[order.TraderID]; !exists {
		return nil, fmt.Errorf("unknown trader: %s", order.TraderID)
	}

	order.ID = uuid.New()
	order.Status = domain.OrderStatusPending
	order.FilledSize = decimal.Zero
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	trades := me.matchOrder(book, order)

	// If order has remaining size and is a limit order, rest it
	if order.RemainingSize().IsPositive() && order.Type == domain.OrderTypeLimit {
		book.AddOrder(order)
		order.Status = domain.OrderStatusPartial
		if order.FilledSize.IsZero() {
			order.Status = domain.OrderStatusPending
		}
		// Persist resting order
		if me.db != nil {
			if err := me.db.SaveOrder(order); err != nil {
				log.Printf("Error saving order to database: %v", err)
			}
		}
	} else if order.RemainingSize().IsZero() {
		order.Status = domain.OrderStatusFilled
	}

	// Notify handlers
	for _, handler := range me.orderHandlers {
		handler(order)
	}

	return trades, nil
}

// matchOrder attempts to match an incoming order against the book
func (me *MatchingEngine) matchOrder(book *OrderBook, order *domain.Order) []*domain.Trade {
	var trades []*domain.Trade
	var matchLevels []*priceLevel

	if order.Side == domain.SideBuy {
		if order.Type == domain.OrderTypeMarket {
			// Market buy matches any ask
			matchLevels = book.matchableAsks(decimal.New(1, 18)) // Very high price
		} else {
			// Limit buy matches asks at or below limit price
			matchLevels = book.matchableAsks(order.Price)
		}
	} else {
		if order.Type == domain.OrderTypeMarket {
			// Market sell matches any bid
			matchLevels = book.matchableBids(decimal.Zero)
		} else {
			// Limit sell matches bids at or above limit price
			matchLevels = book.matchableBids(order.Price)
		}
	}

	for _, level := range matchLevels {
		if order.RemainingSize().IsZero() {
			break
		}

		curr := level.head
		for curr != nil && order.RemainingSize().IsPositive() {
			restingOrder := curr.order

			// Don't self-trade
			if restingOrder.TraderID == order.TraderID {
				curr = curr.next
				continue
			}

			// Calculate fill size
			fillSize := decimal.Min(order.RemainingSize(), restingOrder.RemainingSize())
			fillPrice := restingOrder.Price // Price-time priority: resting order's price

			// Create the trade
			trade := me.createTrade(order, restingOrder, fillPrice, fillSize)
			trades = append(trades, trade)

			// Update order fill sizes
			order.FilledSize = order.FilledSize.Add(fillSize)
			restingOrder.FilledSize = restingOrder.FilledSize.Add(fillSize)
			order.UpdatedAt = time.Now()
			restingOrder.UpdatedAt = time.Now()

			// Update resting order status
			if restingOrder.RemainingSize().IsZero() {
				restingOrder.Status = domain.OrderStatusFilled
				book.RemoveOrder(restingOrder.ID)
				// Remove filled order from database
				if me.db != nil {
					if err := me.db.DeleteOrder(restingOrder.ID); err != nil {
						log.Printf("Error deleting filled order from database: %v", err)
					}
				}
			} else {
				restingOrder.Status = domain.OrderStatusPartial
				// Update level size
				level.totalSize = level.totalSize.Sub(fillSize)
				// Update partial fill in database
				if me.db != nil {
					if err := me.db.SaveOrder(restingOrder); err != nil {
						log.Printf("Error updating order in database: %v", err)
					}
				}
			}

			// Notify about resting order update
			for _, handler := range me.orderHandlers {
				handler(restingOrder)
			}

			// Notify about trade
			for _, handler := range me.tradeHandlers {
				handler(trade)
			}

			curr = curr.next
		}
	}

	return trades
}

// createTrade creates a trade record with full transparency
func (me *MatchingEngine) createTrade(aggressor, resting *domain.Order, price, size decimal.Decimal) *domain.Trade {
	var buyerOrder, sellerOrder *domain.Order
	var aggressorSide domain.Side

	if aggressor.Side == domain.SideBuy {
		buyerOrder = aggressor
		sellerOrder = resting
		aggressorSide = domain.SideBuy
	} else {
		buyerOrder = resting
		sellerOrder = aggressor
		aggressorSide = domain.SideSell
	}

	// Determine position effects
	buyerEffect := me.determinePositionEffect(buyerOrder.TraderID, buyerOrder.Instrument, size)
	sellerEffect := me.determinePositionEffect(sellerOrder.TraderID, sellerOrder.Instrument, size.Neg())

	// Update positions
	buyerNewPos := me.updatePosition(buyerOrder.TraderID, buyerOrder.Instrument, size, price)
	sellerNewPos := me.updatePosition(sellerOrder.TraderID, sellerOrder.Instrument, size.Neg(), price)

	trade := &domain.Trade{
		ID:                uuid.New(),
		Instrument:        aggressor.Instrument,
		Price:             price,
		Size:              size,
		Timestamp:         time.Now(),
		BuyerID:           buyerOrder.TraderID,
		SellerID:          sellerOrder.TraderID,
		BuyerOrderID:      buyerOrder.ID,
		SellerOrderID:     sellerOrder.ID,
		BuyerEffect:       buyerEffect,
		SellerEffect:      sellerEffect,
		BuyerNewPosition:  buyerNewPos,
		SellerNewPosition: sellerNewPos,
		AggressorSide:     aggressorSide,
	}

	// Update trader stats
	if buyer, ok := me.traders[buyerOrder.TraderID]; ok {
		buyer.TradeCount++
	}
	if seller, ok := me.traders[sellerOrder.TraderID]; ok {
		seller.TradeCount++
	}

	// Store trade in history (keep last 1000)
	me.recentTrades = append([]*domain.Trade{trade}, me.recentTrades...)
	if len(me.recentTrades) > 1000 {
		me.recentTrades = me.recentTrades[:1000]
	}

	// Persist to database
	if me.db != nil {
		if err := me.db.SaveTrade(trade); err != nil {
			log.Printf("Error saving trade to database: %v", err)
		}
		// Save updated trader stats
		if buyer, ok := me.traders[buyerOrder.TraderID]; ok {
			if err := me.db.SaveTrader(buyer); err != nil {
				log.Printf("Error saving buyer to database: %v", err)
			}
		}
		if seller, ok := me.traders[sellerOrder.TraderID]; ok {
			if err := me.db.SaveTrader(seller); err != nil {
				log.Printf("Error saving seller to database: %v", err)
			}
		}
	}

	return trade
}

// determinePositionEffect figures out what this trade does to the position
func (me *MatchingEngine) determinePositionEffect(traderID uuid.UUID, instrument string, sizeChange decimal.Decimal) domain.PositionEffect {
	posKey := fmt.Sprintf("%s:%s", traderID, instrument)
	pos, exists := me.positions[posKey]

	if !exists || pos.Size.IsZero() {
		return domain.EffectOpen
	}

	// If signs match, it's adding to position (opening more)
	if (pos.Size.IsPositive() && sizeChange.IsPositive()) ||
		(pos.Size.IsNegative() && sizeChange.IsNegative()) {
		return domain.EffectOpen
	}

	// Signs differ, it's closing
	return domain.EffectClose
}

// updatePosition updates a trader's position and returns new size
func (me *MatchingEngine) updatePosition(traderID uuid.UUID, instrument string, sizeChange, price decimal.Decimal) decimal.Decimal {
	posKey := fmt.Sprintf("%s:%s", traderID, instrument)
	pos, exists := me.positions[posKey]

	if !exists {
		pos = &domain.Position{
			TraderID:      traderID,
			Instrument:    instrument,
			Size:          decimal.Zero,
			EntryPrice:    decimal.Zero,
			UnrealizedPnL: decimal.Zero,
			RealizedPnL:   decimal.Zero,
			Leverage:      1,
		}
		me.positions[posKey] = pos
	}

	oldSize := pos.Size
	newSize := oldSize.Add(sizeChange)

	// Calculate new entry price (weighted average for opening, unchanged for closing)
	if oldSize.IsZero() {
		pos.EntryPrice = price
	} else if (oldSize.IsPositive() && sizeChange.IsPositive()) ||
		(oldSize.IsNegative() && sizeChange.IsNegative()) {
		// Adding to position - weighted average
		totalCost := oldSize.Mul(pos.EntryPrice).Add(sizeChange.Mul(price))
		pos.EntryPrice = totalCost.Div(newSize)
	} else {
		// Reducing position - realize P&L
		closedSize := decimal.Min(oldSize.Abs(), sizeChange.Abs())
		if oldSize.IsPositive() {
			// Was long, selling - profit if price > entry
			pnl := price.Sub(pos.EntryPrice).Mul(closedSize)
			pos.RealizedPnL = pos.RealizedPnL.Add(pnl)
		} else {
			// Was short, buying - profit if price < entry
			pnl := pos.EntryPrice.Sub(price).Mul(closedSize)
			pos.RealizedPnL = pos.RealizedPnL.Add(pnl)
		}

		// If flipping sides, set new entry for the overflow
		if !newSize.IsZero() && ((oldSize.IsPositive() && newSize.IsNegative()) ||
			(oldSize.IsNegative() && newSize.IsPositive())) {
			pos.EntryPrice = price
		}
	}

	pos.Size = newSize
	pos.UpdatedAt = time.Now()

	// Calculate liquidation price if position exists
	if !newSize.IsZero() {
		pos.LiquidationPrice = me.calculateLiquidationPrice(pos.EntryPrice, pos.Leverage, newSize.IsPositive())
	}

	// Persist position to database
	if me.db != nil {
		if newSize.IsZero() {
			// Position closed, delete from database
			if err := me.db.DeletePosition(traderID, instrument); err != nil {
				log.Printf("Error deleting position from database: %v", err)
			}
		} else {
			if err := me.db.SavePosition(pos); err != nil {
				log.Printf("Error saving position to database: %v", err)
			}
		}
	}

	return newSize
}

// GetPosition returns a trader's position (public - transparency!)
func (me *MatchingEngine) GetPosition(traderID uuid.UUID, instrument string) *domain.Position {
	me.mu.RLock()
	defer me.mu.RUnlock()

	posKey := fmt.Sprintf("%s:%s", traderID, instrument)
	pos, exists := me.positions[posKey]
	if !exists {
		return nil
	}
	return pos
}

// GetAllPositions returns all positions for an instrument (transparency!)
func (me *MatchingEngine) GetAllPositions(instrument string) []*domain.Position {
	me.mu.RLock()
	defer me.mu.RUnlock()

	var positions []*domain.Position
	for key, pos := range me.positions {
		if pos.Instrument == instrument && !pos.Size.IsZero() {
			_ = key
			positions = append(positions, pos)
		}
	}
	return positions
}

// GetOrderBook returns the order book for an instrument
func (me *MatchingEngine) GetOrderBook(instrument string, depth int) (*domain.OrderBook, error) {
	me.mu.RLock()
	defer me.mu.RUnlock()

	book, exists := me.books[instrument]
	if !exists {
		return nil, fmt.Errorf("unknown instrument: %s", instrument)
	}

	snapshot := book.GetSnapshot(depth)
	return &snapshot, nil
}

// CancelOrder cancels an existing order
func (me *MatchingEngine) CancelOrder(orderID uuid.UUID, instrument string) error {
	me.mu.Lock()
	defer me.mu.Unlock()

	book, exists := me.books[instrument]
	if !exists {
		return fmt.Errorf("unknown instrument: %s", instrument)
	}

	order, exists := book.GetOrder(orderID)
	if !exists {
		return fmt.Errorf("order not found: %s", orderID)
	}

	book.RemoveOrder(orderID)
	order.Status = domain.OrderStatusCancelled
	order.UpdatedAt = time.Now()

	// Remove from database
	if me.db != nil {
		if err := me.db.DeleteOrder(orderID); err != nil {
			log.Printf("Error deleting order from database: %v", err)
		}
	}

	for _, handler := range me.orderHandlers {
		handler(order)
	}

	return nil
}

// GetOpenInterestBreakdown calculates OI stats (the core transparency feature!)
func (me *MatchingEngine) GetOpenInterestBreakdown(instrument string) *domain.OpenInterestBreakdown {
	me.mu.RLock()
	defer me.mu.RUnlock()

	breakdown := &domain.OpenInterestBreakdown{
		Instrument: instrument,
		Timestamp:  time.Now(),
	}

	for _, pos := range me.positions {
		if pos.Instrument != instrument || pos.Size.IsZero() {
			continue
		}

		if pos.Size.IsPositive() {
			breakdown.LongPositions++
			breakdown.TotalOI = breakdown.TotalOI.Add(pos.Size)
		} else {
			breakdown.ShortPositions++
		}
	}

	return breakdown
}

// GetTrader returns trader info (public)
func (me *MatchingEngine) GetTrader(traderID uuid.UUID) *domain.Trader {
	me.mu.RLock()
	defer me.mu.RUnlock()
	return me.traders[traderID]
}

// GetAllTraders returns all traders (public)
func (me *MatchingEngine) GetAllTraders() []*domain.Trader {
	me.mu.RLock()
	defer me.mu.RUnlock()

	traders := make([]*domain.Trader, 0, len(me.traders))
	for _, t := range me.traders {
		traders = append(traders, t)
	}
	return traders
}

// GetRecentTrades returns recent trades for an instrument
func (me *MatchingEngine) GetRecentTrades(instrument string, limit int) []*domain.Trade {
	me.mu.RLock()
	defer me.mu.RUnlock()

	var trades []*domain.Trade
	for _, t := range me.recentTrades {
		if t.Instrument == instrument {
			trades = append(trades, t)
			if len(trades) >= limit {
				break
			}
		}
	}
	return trades
}

// GetTraderTrades returns trades where the trader was buyer or seller
func (me *MatchingEngine) GetTraderTrades(traderID uuid.UUID, instrument string, limit int) []*domain.Trade {
	me.mu.RLock()
	defer me.mu.RUnlock()

	var trades []*domain.Trade
	for _, t := range me.recentTrades {
		if t.Instrument == instrument && (t.BuyerID == traderID || t.SellerID == traderID) {
			trades = append(trades, t)
			if len(trades) >= limit {
				break
			}
		}
	}
	return trades
}

// GetRecentLiquidations returns recent liquidations for an instrument
func (me *MatchingEngine) GetRecentLiquidations(instrument string, limit int) []*domain.Liquidation {
	me.mu.RLock()
	defer me.mu.RUnlock()

	var liqs []*domain.Liquidation
	for _, l := range me.liquidations {
		if l.Instrument == instrument {
			liqs = append(liqs, l)
			if len(liqs) >= limit {
				break
			}
		}
	}
	return liqs
}

// GetCandles returns OHLCV candles for an instrument
func (me *MatchingEngine) GetCandles(instrument string, interval domain.CandleInterval, limit int) []*domain.Candle {
	me.mu.RLock()
	defer me.mu.RUnlock()

	// Get interval duration
	intervalDuration := getIntervalDuration(interval)

	// Group trades by candle period
	candleMap := make(map[int64]*domain.Candle)

	for _, t := range me.recentTrades {
		if t.Instrument != instrument {
			continue
		}

		// Calculate candle start time (truncate to interval)
		candleStart := truncateToInterval(t.Timestamp, intervalDuration)
		candleKey := candleStart.Unix()

		candle, exists := candleMap[candleKey]
		if !exists {
			candle = &domain.Candle{
				Instrument: instrument,
				Interval:   interval,
				OpenTime:   candleStart,
				CloseTime:  candleStart.Add(intervalDuration),
				Open:       t.Price,
				High:       t.Price,
				Low:        t.Price,
				Close:      t.Price,
				Volume:     t.Size,
				TradeCount: 1,
			}
			candleMap[candleKey] = candle
		} else {
			// Update OHLCV - trades are newest first, so this trade is older
			candle.Open = t.Price // Keep updating open since we iterate newest->oldest
			if t.Price.GreaterThan(candle.High) {
				candle.High = t.Price
			}
			if t.Price.LessThan(candle.Low) {
				candle.Low = t.Price
			}
			candle.Volume = candle.Volume.Add(t.Size)
			candle.TradeCount++
		}
	}

	// Convert map to sorted slice (newest first)
	var candles []*domain.Candle
	for _, c := range candleMap {
		candles = append(candles, c)
	}

	// Sort by open time descending (newest first)
	for i := 0; i < len(candles)-1; i++ {
		for j := i + 1; j < len(candles); j++ {
			if candles[j].OpenTime.After(candles[i].OpenTime) {
				candles[i], candles[j] = candles[j], candles[i]
			}
		}
	}

	// Limit results
	if len(candles) > limit {
		candles = candles[:limit]
	}

	return candles
}

// GetHistoricalTrades returns trades within a time range
func (me *MatchingEngine) GetHistoricalTrades(instrument string, start, end time.Time, limit int) []*domain.Trade {
	me.mu.RLock()
	defer me.mu.RUnlock()

	var trades []*domain.Trade
	for _, t := range me.recentTrades {
		if t.Instrument != instrument {
			continue
		}
		if t.Timestamp.Before(start) || t.Timestamp.After(end) {
			continue
		}
		trades = append(trades, t)
		if len(trades) >= limit {
			break
		}
	}
	return trades
}

// GetHistoricalCandles returns candles within a time range
func (me *MatchingEngine) GetHistoricalCandles(instrument string, interval domain.CandleInterval, start, end time.Time, limit int) []*domain.Candle {
	me.mu.RLock()
	defer me.mu.RUnlock()

	intervalDuration := getIntervalDuration(interval)
	candleMap := make(map[int64]*domain.Candle)

	for _, t := range me.recentTrades {
		if t.Instrument != instrument {
			continue
		}
		if t.Timestamp.Before(start) || t.Timestamp.After(end) {
			continue
		}

		candleStart := truncateToInterval(t.Timestamp, intervalDuration)
		candleKey := candleStart.Unix()

		candle, exists := candleMap[candleKey]
		if !exists {
			candle = &domain.Candle{
				Instrument: instrument,
				Interval:   interval,
				OpenTime:   candleStart,
				CloseTime:  candleStart.Add(intervalDuration),
				Open:       t.Price,
				High:       t.Price,
				Low:        t.Price,
				Close:      t.Price,
				Volume:     t.Size,
				TradeCount: 1,
			}
			candleMap[candleKey] = candle
		} else {
			candle.Open = t.Price
			if t.Price.GreaterThan(candle.High) {
				candle.High = t.Price
			}
			if t.Price.LessThan(candle.Low) {
				candle.Low = t.Price
			}
			candle.Volume = candle.Volume.Add(t.Size)
			candle.TradeCount++
		}
	}

	var candles []*domain.Candle
	for _, c := range candleMap {
		candles = append(candles, c)
	}

	// Sort by open time ascending (oldest first for historical)
	for i := 0; i < len(candles)-1; i++ {
		for j := i + 1; j < len(candles); j++ {
			if candles[j].OpenTime.Before(candles[i].OpenTime) {
				candles[i], candles[j] = candles[j], candles[i]
			}
		}
	}

	if len(candles) > limit {
		candles = candles[:limit]
	}

	return candles
}

// getIntervalDuration converts interval string to duration
func getIntervalDuration(interval domain.CandleInterval) time.Duration {
	switch interval {
	case domain.CandleInterval1m:
		return time.Minute
	case domain.CandleInterval5m:
		return 5 * time.Minute
	case domain.CandleInterval15m:
		return 15 * time.Minute
	case domain.CandleInterval1h:
		return time.Hour
	case domain.CandleInterval4h:
		return 4 * time.Hour
	case domain.CandleInterval1d:
		return 24 * time.Hour
	default:
		return time.Minute
	}
}

// truncateToInterval truncates time to interval boundary
func truncateToInterval(t time.Time, d time.Duration) time.Time {
	return t.UTC().Truncate(d)
}

// GetMarketStats returns market statistics for an instrument
func (me *MatchingEngine) GetMarketStats(instrument string) *domain.MarketStats {
	me.mu.RLock()
	defer me.mu.RUnlock()

	stats := &domain.MarketStats{
		Instrument:    instrument,
		Timestamp:     time.Now(),
		InsuranceFund: decimal.NewFromInt(1000000), // Default
	}

	// Get last price from recent trades
	for _, t := range me.recentTrades {
		if t.Instrument == instrument {
			stats.LastPrice = t.Price
			stats.MarkPrice = t.Price
			break
		}
	}

	// If no trades yet, use 1000 as starting price
	if stats.LastPrice.IsZero() {
		stats.LastPrice = decimal.NewFromInt(1000)
		stats.MarkPrice = decimal.NewFromInt(1000)
	}

	// Calculate 24h stats from trades
	oneDayAgo := time.Now().Add(-24 * time.Hour)
	stats.High24h = stats.LastPrice
	stats.Low24h = stats.LastPrice

	for _, t := range me.recentTrades {
		if t.Instrument == instrument && t.Timestamp.After(oneDayAgo) {
			if t.Price.GreaterThan(stats.High24h) {
				stats.High24h = t.Price
			}
			if t.Price.LessThan(stats.Low24h) {
				stats.Low24h = t.Price
			}
			stats.Volume24h = stats.Volume24h.Add(t.Size.Mul(t.Price))
		}
	}

	// Calculate open interest
	for _, pos := range me.positions {
		if pos.Instrument == instrument && !pos.Size.IsZero() {
			stats.OpenInterest = stats.OpenInterest.Add(pos.Size.Abs())
		}
	}

	return stats
}

// SetLiquidationConfig sets the liquidation configuration for calculating liquidation prices
func (me *MatchingEngine) SetLiquidationConfig(cfg *config.LiquidationConfig) {
	me.liqConfig = cfg
}

// GetMarkPrice returns the current mark price for an instrument (implements PriceProvider)
func (me *MatchingEngine) GetMarkPrice(instrument string) decimal.Decimal {
	me.mu.RLock()
	defer me.mu.RUnlock()

	// Get last trade price as mark price
	for _, t := range me.recentTrades {
		if t.Instrument == instrument {
			return t.Price
		}
	}

	// Default to 1000 if no trades
	return decimal.NewFromInt(1000)
}

// ClosePosition closes a position at the given mark price (implements PositionStore)
func (me *MatchingEngine) ClosePosition(traderID uuid.UUID, instrument string, markPrice decimal.Decimal) error {
	me.mu.Lock()
	defer me.mu.Unlock()

	posKey := fmt.Sprintf("%s:%s", traderID, instrument)
	pos, exists := me.positions[posKey]
	if !exists || pos.Size.IsZero() {
		return fmt.Errorf("no position to close")
	}

	// Calculate realized P&L
	var pnl decimal.Decimal
	if pos.IsLong() {
		pnl = markPrice.Sub(pos.EntryPrice).Mul(pos.Size)
	} else {
		pnl = pos.EntryPrice.Sub(markPrice).Mul(pos.Size.Abs())
	}

	// Update trader balance
	if trader, ok := me.traders[traderID]; ok {
		trader.Balance = trader.Balance.Add(pos.Margin).Add(pnl)
		trader.TotalPnL = trader.TotalPnL.Add(pnl)
		if me.db != nil {
			if err := me.db.SaveTrader(trader); err != nil {
				log.Printf("Error saving trader after liquidation: %v", err)
			}
		}
	}

	// Delete position
	delete(me.positions, posKey)
	if me.db != nil {
		if err := me.db.DeletePosition(traderID, instrument); err != nil {
			log.Printf("Error deleting liquidated position: %v", err)
		}
	}

	return nil
}

// OnLiquidation registers a liquidation handler
func (me *MatchingEngine) OnLiquidation(handler LiquidationHandler) {
	me.liquidationHandlers = append(me.liquidationHandlers, handler)
}

// AddLiquidation adds a liquidation to history and notifies handlers
func (me *MatchingEngine) AddLiquidation(liq *domain.Liquidation) {
	me.mu.Lock()
	defer me.mu.Unlock()

	// Add to history
	me.liquidations = append([]*domain.Liquidation{liq}, me.liquidations...)
	if len(me.liquidations) > 100 {
		me.liquidations = me.liquidations[:100]
	}

	// Persist to database
	if me.db != nil {
		if err := me.db.SaveLiquidation(liq); err != nil {
			log.Printf("Error saving liquidation: %v", err)
		}
	}

	// Notify handlers
	for _, handler := range me.liquidationHandlers {
		handler(liq)
	}
}

// calculateLiquidationPrice computes liquidation price for a position
func (me *MatchingEngine) calculateLiquidationPrice(entryPrice decimal.Decimal, leverage int, isLong bool) decimal.Decimal {
	if me.liqConfig == nil {
		// Default maintenance margins if not configured
		return decimal.Zero
	}

	maintMargin := me.liqConfig.MaintenanceMargins.GetMarginForLeverage(leverage)
	leverageDecimal := decimal.NewFromInt(int64(leverage))

	// Liquidation distance = entry / leverage * (1 - maintenance margin)
	distance := entryPrice.Div(leverageDecimal).Mul(decimal.NewFromInt(1).Sub(maintMargin))

	if isLong {
		return entryPrice.Sub(distance)
	}
	return entryPrice.Add(distance)
}
