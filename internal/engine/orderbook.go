package engine

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/thatreguy/trade.re/internal/domain"
)

// orderNode represents an order in the queue at a price level
type orderNode struct {
	order *domain.Order
	next  *orderNode
}

// priceLevel represents all orders at a specific price
type priceLevel struct {
	price      decimal.Decimal
	totalSize  decimal.Decimal
	head       *orderNode
	tail       *orderNode
	orderCount int
}

// OrderBook manages buy and sell orders for an instrument
type OrderBook struct {
	instrument string
	bids       map[string]*priceLevel // price string -> level (buys)
	asks       map[string]*priceLevel // price string -> level (sells)
	orders     map[uuid.UUID]*domain.Order // quick order lookup
	mu         sync.RWMutex
}

// NewOrderBook creates a new order book for an instrument
func NewOrderBook(instrument string) *OrderBook {
	return &OrderBook{
		instrument: instrument,
		bids:       make(map[string]*priceLevel),
		asks:       make(map[string]*priceLevel),
		orders:     make(map[uuid.UUID]*domain.Order),
	}
}

// AddOrder adds an order to the book (does not match, just rests)
func (ob *OrderBook) AddOrder(order *domain.Order) {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	priceKey := order.Price.String()
	var levels map[string]*priceLevel

	if order.Side == domain.SideBuy {
		levels = ob.bids
	} else {
		levels = ob.asks
	}

	level, exists := levels[priceKey]
	if !exists {
		level = &priceLevel{
			price:     order.Price,
			totalSize: decimal.Zero,
		}
		levels[priceKey] = level
	}

	// Add to FIFO queue
	node := &orderNode{order: order}
	if level.tail == nil {
		level.head = node
		level.tail = node
	} else {
		level.tail.next = node
		level.tail = node
	}

	level.totalSize = level.totalSize.Add(order.RemainingSize())
	level.orderCount++
	ob.orders[order.ID] = order
}

// RemoveOrder removes an order from the book
func (ob *OrderBook) RemoveOrder(orderID uuid.UUID) bool {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	order, exists := ob.orders[orderID]
	if !exists {
		return false
	}

	priceKey := order.Price.String()
	var levels map[string]*priceLevel

	if order.Side == domain.SideBuy {
		levels = ob.bids
	} else {
		levels = ob.asks
	}

	level, exists := levels[priceKey]
	if !exists {
		return false
	}

	// Remove from linked list
	var prev *orderNode
	curr := level.head
	for curr != nil {
		if curr.order.ID == orderID {
			if prev == nil {
				level.head = curr.next
			} else {
				prev.next = curr.next
			}
			if curr == level.tail {
				level.tail = prev
			}
			level.totalSize = level.totalSize.Sub(order.RemainingSize())
			level.orderCount--
			break
		}
		prev = curr
		curr = curr.next
	}

	// Remove empty price level
	if level.orderCount == 0 {
		delete(levels, priceKey)
	}

	delete(ob.orders, orderID)
	return true
}

// GetOrder retrieves an order by ID
func (ob *OrderBook) GetOrder(orderID uuid.UUID) (*domain.Order, bool) {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	order, exists := ob.orders[orderID]
	return order, exists
}

// BestBid returns the highest bid price and size
func (ob *OrderBook) BestBid() (decimal.Decimal, decimal.Decimal, bool) {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	var bestPrice decimal.Decimal
	var bestLevel *priceLevel
	first := true

	for _, level := range ob.bids {
		if first || level.price.GreaterThan(bestPrice) {
			bestPrice = level.price
			bestLevel = level
			first = false
		}
	}

	if bestLevel == nil {
		return decimal.Zero, decimal.Zero, false
	}
	return bestPrice, bestLevel.totalSize, true
}

// BestAsk returns the lowest ask price and size
func (ob *OrderBook) BestAsk() (decimal.Decimal, decimal.Decimal, bool) {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	var bestPrice decimal.Decimal
	var bestLevel *priceLevel
	first := true

	for _, level := range ob.asks {
		if first || level.price.LessThan(bestPrice) {
			bestPrice = level.price
			bestLevel = level
			first = false
		}
	}

	if bestLevel == nil {
		return decimal.Zero, decimal.Zero, false
	}
	return bestPrice, bestLevel.totalSize, true
}

// GetSnapshot returns the current order book state
func (ob *OrderBook) GetSnapshot(depth int) domain.OrderBook {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	snapshot := domain.OrderBook{
		Instrument: ob.instrument,
		Timestamp:  time.Now(),
		Bids:       make([]domain.OrderBookLevel, 0, depth),
		Asks:       make([]domain.OrderBookLevel, 0, depth),
	}

	// Collect and sort bids (highest first)
	bidLevels := make([]*priceLevel, 0, len(ob.bids))
	for _, level := range ob.bids {
		bidLevels = append(bidLevels, level)
	}
	sortLevelsDesc(bidLevels)

	for i, level := range bidLevels {
		if i >= depth {
			break
		}
		snapshot.Bids = append(snapshot.Bids, domain.OrderBookLevel{
			Price:      level.price,
			Size:       level.totalSize,
			OrderCount: level.orderCount,
		})
	}

	// Collect and sort asks (lowest first)
	askLevels := make([]*priceLevel, 0, len(ob.asks))
	for _, level := range ob.asks {
		askLevels = append(askLevels, level)
	}
	sortLevelsAsc(askLevels)

	for i, level := range askLevels {
		if i >= depth {
			break
		}
		snapshot.Asks = append(snapshot.Asks, domain.OrderBookLevel{
			Price:      level.price,
			Size:       level.totalSize,
			OrderCount: level.orderCount,
		})
	}

	return snapshot
}

// GetOrdersAtPrice returns all orders at a price level (for transparency)
func (ob *OrderBook) GetOrdersAtPrice(side domain.Side, price decimal.Decimal) []*domain.Order {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	priceKey := price.String()
	var levels map[string]*priceLevel

	if side == domain.SideBuy {
		levels = ob.bids
	} else {
		levels = ob.asks
	}

	level, exists := levels[priceKey]
	if !exists {
		return nil
	}

	orders := make([]*domain.Order, 0, level.orderCount)
	curr := level.head
	for curr != nil {
		orders = append(orders, curr.order)
		curr = curr.next
	}
	return orders
}

// matchableBids returns bid levels that can match at or above the given price
func (ob *OrderBook) matchableBids(price decimal.Decimal) []*priceLevel {
	levels := make([]*priceLevel, 0)
	for _, level := range ob.bids {
		if level.price.GreaterThanOrEqual(price) {
			levels = append(levels, level)
		}
	}
	sortLevelsDesc(levels) // Best price first
	return levels
}

// matchableAsks returns ask levels that can match at or below the given price
func (ob *OrderBook) matchableAsks(price decimal.Decimal) []*priceLevel {
	levels := make([]*priceLevel, 0)
	for _, level := range ob.asks {
		if level.price.LessThanOrEqual(price) {
			levels = append(levels, level)
		}
	}
	sortLevelsAsc(levels) // Best price first
	return levels
}

// Helper sort functions
func sortLevelsDesc(levels []*priceLevel) {
	for i := 0; i < len(levels)-1; i++ {
		for j := i + 1; j < len(levels); j++ {
			if levels[j].price.GreaterThan(levels[i].price) {
				levels[i], levels[j] = levels[j], levels[i]
			}
		}
	}
}

func sortLevelsAsc(levels []*priceLevel) {
	for i := 0; i < len(levels)-1; i++ {
		for j := i + 1; j < len(levels); j++ {
			if levels[j].price.LessThan(levels[i].price) {
				levels[i], levels[j] = levels[j], levels[i]
			}
		}
	}
}
