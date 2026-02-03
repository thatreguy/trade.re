package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/thatreguy/trade.re/internal/domain"
	_ "modernc.org/sqlite"
)

// SQLiteDB wraps the SQLite connection
type SQLiteDB struct {
	db *sql.DB
}

// NewSQLite creates a new SQLite database connection
func NewSQLite(dbPath string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Enable WAL mode for better concurrent access
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("setting WAL mode: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return nil, fmt.Errorf("enabling foreign keys: %w", err)
	}

	sqlite := &SQLiteDB{db: db}

	// Create tables
	if err := sqlite.createTables(); err != nil {
		return nil, fmt.Errorf("creating tables: %w", err)
	}

	return sqlite, nil
}

// createTables creates the database schema
func (s *SQLiteDB) createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS traders (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL DEFAULT '',
		type TEXT NOT NULL DEFAULT 'human',
		balance TEXT NOT NULL DEFAULT '10000',
		total_pnl TEXT NOT NULL DEFAULT '0',
		trade_count INTEGER NOT NULL DEFAULT 0,
		max_leverage_used INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS positions (
		trader_id TEXT NOT NULL,
		instrument TEXT NOT NULL,
		size TEXT NOT NULL,
		entry_price TEXT NOT NULL,
		leverage INTEGER NOT NULL DEFAULT 1,
		margin TEXT NOT NULL DEFAULT '0',
		unrealized_pnl TEXT NOT NULL DEFAULT '0',
		realized_pnl TEXT NOT NULL DEFAULT '0',
		liquidation_price TEXT NOT NULL DEFAULT '0',
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (trader_id) REFERENCES traders(id),
		PRIMARY KEY(trader_id, instrument)
	);

	CREATE TABLE IF NOT EXISTS orders (
		id TEXT PRIMARY KEY,
		trader_id TEXT NOT NULL,
		instrument TEXT NOT NULL,
		side TEXT NOT NULL,
		type TEXT NOT NULL,
		price TEXT NOT NULL,
		size TEXT NOT NULL,
		filled_size TEXT NOT NULL DEFAULT '0',
		status TEXT NOT NULL DEFAULT 'open',
		leverage INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (trader_id) REFERENCES traders(id)
	);

	CREATE TABLE IF NOT EXISTS trades (
		id TEXT PRIMARY KEY,
		instrument TEXT NOT NULL,
		price TEXT NOT NULL,
		size TEXT NOT NULL,
		buyer_id TEXT NOT NULL,
		seller_id TEXT NOT NULL,
		buyer_leverage INTEGER NOT NULL DEFAULT 1,
		seller_leverage INTEGER NOT NULL DEFAULT 1,
		buyer_effect TEXT NOT NULL DEFAULT 'open',
		seller_effect TEXT NOT NULL DEFAULT 'open',
		aggressor_side TEXT NOT NULL,
		timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (buyer_id) REFERENCES traders(id),
		FOREIGN KEY (seller_id) REFERENCES traders(id)
	);

	CREATE TABLE IF NOT EXISTS liquidations (
		id TEXT PRIMARY KEY,
		trader_id TEXT NOT NULL,
		instrument TEXT NOT NULL,
		side TEXT NOT NULL,
		size TEXT NOT NULL,
		entry_price TEXT NOT NULL,
		liquidation_price TEXT NOT NULL,
		mark_price TEXT NOT NULL,
		leverage INTEGER NOT NULL,
		loss TEXT NOT NULL,
		insurance_fund_hit INTEGER NOT NULL DEFAULT 0,
		timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (trader_id) REFERENCES traders(id)
	);

	CREATE TABLE IF NOT EXISTS market_stats (
		instrument TEXT PRIMARY KEY,
		last_price TEXT NOT NULL DEFAULT '1000',
		mark_price TEXT NOT NULL DEFAULT '1000',
		high_24h TEXT NOT NULL DEFAULT '0',
		low_24h TEXT NOT NULL DEFAULT '0',
		volume_24h TEXT NOT NULL DEFAULT '0',
		open_interest TEXT NOT NULL DEFAULT '0',
		insurance_fund TEXT NOT NULL DEFAULT '1000000',
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_positions_trader ON positions(trader_id);
	CREATE INDEX IF NOT EXISTS idx_orders_trader ON orders(trader_id);
	CREATE INDEX IF NOT EXISTS idx_orders_instrument_status ON orders(instrument, status);
	CREATE INDEX IF NOT EXISTS idx_trades_instrument ON trades(instrument);
	CREATE INDEX IF NOT EXISTS idx_trades_timestamp ON trades(timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_trades_buyer ON trades(buyer_id);
	CREATE INDEX IF NOT EXISTS idx_trades_seller ON trades(seller_id);
	CREATE INDEX IF NOT EXISTS idx_liquidations_instrument ON liquidations(instrument);
	`

	_, err := s.db.Exec(schema)
	return err
}

// Close closes the database connection
func (s *SQLiteDB) Close() error {
	return s.db.Close()
}

// === Trader Operations ===

// SaveTrader inserts or updates a trader
func (s *SQLiteDB) SaveTrader(trader *domain.Trader) error {
	query := `
	INSERT INTO traders (id, username, password_hash, type, balance, total_pnl, trade_count, max_leverage_used, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		username = excluded.username,
		balance = excluded.balance,
		total_pnl = excluded.total_pnl,
		trade_count = excluded.trade_count,
		max_leverage_used = excluded.max_leverage_used
	`
	_, err := s.db.Exec(query,
		trader.ID.String(),
		trader.Username,
		trader.PasswordHash,
		string(trader.Type),
		trader.Balance.String(),
		trader.TotalPnL.String(),
		trader.TradeCount,
		trader.MaxLeverageUsed,
		trader.CreatedAt,
	)
	return err
}

// GetTrader retrieves a trader by ID
func (s *SQLiteDB) GetTrader(id uuid.UUID) (*domain.Trader, error) {
	query := `SELECT id, username, password_hash, type, balance, total_pnl, trade_count, max_leverage_used, created_at FROM traders WHERE id = ?`
	row := s.db.QueryRow(query, id.String())

	var trader domain.Trader
	var idStr, typeStr, balanceStr, pnlStr string
	err := row.Scan(&idStr, &trader.Username, &trader.PasswordHash, &typeStr, &balanceStr, &pnlStr, &trader.TradeCount, &trader.MaxLeverageUsed, &trader.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	trader.ID, _ = uuid.Parse(idStr)
	trader.Type = domain.TraderType(typeStr)
	trader.Balance, _ = decimal.NewFromString(balanceStr)
	trader.TotalPnL, _ = decimal.NewFromString(pnlStr)

	return &trader, nil
}

// GetTraderByUsername retrieves a trader by username
func (s *SQLiteDB) GetTraderByUsername(username string) (*domain.Trader, error) {
	query := `SELECT id, username, password_hash, type, balance, total_pnl, trade_count, max_leverage_used, created_at FROM traders WHERE username = ?`
	row := s.db.QueryRow(query, username)

	var trader domain.Trader
	var idStr, typeStr, balanceStr, pnlStr string
	err := row.Scan(&idStr, &trader.Username, &trader.PasswordHash, &typeStr, &balanceStr, &pnlStr, &trader.TradeCount, &trader.MaxLeverageUsed, &trader.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	trader.ID, _ = uuid.Parse(idStr)
	trader.Type = domain.TraderType(typeStr)
	trader.Balance, _ = decimal.NewFromString(balanceStr)
	trader.TotalPnL, _ = decimal.NewFromString(pnlStr)

	return &trader, nil
}

// GetAllTraders retrieves all traders
func (s *SQLiteDB) GetAllTraders() ([]*domain.Trader, error) {
	query := `SELECT id, username, password_hash, type, balance, total_pnl, trade_count, max_leverage_used, created_at FROM traders ORDER BY created_at DESC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var traders []*domain.Trader
	for rows.Next() {
		var trader domain.Trader
		var idStr, typeStr, balanceStr, pnlStr string
		if err := rows.Scan(&idStr, &trader.Username, &trader.PasswordHash, &typeStr, &balanceStr, &pnlStr, &trader.TradeCount, &trader.MaxLeverageUsed, &trader.CreatedAt); err != nil {
			return nil, err
		}
		trader.ID, _ = uuid.Parse(idStr)
		trader.Type = domain.TraderType(typeStr)
		trader.Balance, _ = decimal.NewFromString(balanceStr)
		trader.TotalPnL, _ = decimal.NewFromString(pnlStr)
		traders = append(traders, &trader)
	}

	return traders, nil
}

// === Position Operations ===

// SavePosition inserts or updates a position
func (s *SQLiteDB) SavePosition(pos *domain.Position) error {
	query := `
	INSERT INTO positions (trader_id, instrument, size, entry_price, leverage, margin, unrealized_pnl, realized_pnl, liquidation_price, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(trader_id, instrument) DO UPDATE SET
		size = excluded.size,
		entry_price = excluded.entry_price,
		leverage = excluded.leverage,
		margin = excluded.margin,
		unrealized_pnl = excluded.unrealized_pnl,
		realized_pnl = excluded.realized_pnl,
		liquidation_price = excluded.liquidation_price,
		updated_at = excluded.updated_at
	`
	_, err := s.db.Exec(query,
		pos.TraderID.String(),
		pos.Instrument,
		pos.Size.String(),
		pos.EntryPrice.String(),
		pos.Leverage,
		pos.Margin.String(),
		pos.UnrealizedPnL.String(),
		pos.RealizedPnL.String(),
		pos.LiquidationPrice.String(),
		time.Now(),
	)
	return err
}

// DeletePosition removes a position (when closed)
func (s *SQLiteDB) DeletePosition(traderID uuid.UUID, instrument string) error {
	_, err := s.db.Exec("DELETE FROM positions WHERE trader_id = ? AND instrument = ?", traderID.String(), instrument)
	return err
}

// GetPosition retrieves a position
func (s *SQLiteDB) GetPosition(traderID uuid.UUID, instrument string) (*domain.Position, error) {
	query := `SELECT trader_id, instrument, size, entry_price, leverage, margin, unrealized_pnl, realized_pnl, liquidation_price, updated_at FROM positions WHERE trader_id = ? AND instrument = ?`
	row := s.db.QueryRow(query, traderID.String(), instrument)

	var pos domain.Position
	var traderIDStr, sizeStr, entryStr, marginStr, unrealizedStr, realizedStr, liqStr string
	err := row.Scan(&traderIDStr, &pos.Instrument, &sizeStr, &entryStr, &pos.Leverage, &marginStr, &unrealizedStr, &realizedStr, &liqStr, &pos.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	pos.TraderID, _ = uuid.Parse(traderIDStr)
	pos.Size, _ = decimal.NewFromString(sizeStr)
	pos.EntryPrice, _ = decimal.NewFromString(entryStr)
	pos.Margin, _ = decimal.NewFromString(marginStr)
	pos.UnrealizedPnL, _ = decimal.NewFromString(unrealizedStr)
	pos.RealizedPnL, _ = decimal.NewFromString(realizedStr)
	pos.LiquidationPrice, _ = decimal.NewFromString(liqStr)

	return &pos, nil
}

// GetAllPositions retrieves all positions for an instrument
func (s *SQLiteDB) GetAllPositions(instrument string) ([]*domain.Position, error) {
	query := `SELECT trader_id, instrument, size, entry_price, leverage, margin, unrealized_pnl, realized_pnl, liquidation_price, updated_at FROM positions WHERE instrument = ?`
	rows, err := s.db.Query(query, instrument)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []*domain.Position
	for rows.Next() {
		var pos domain.Position
		var traderIDStr, sizeStr, entryStr, marginStr, unrealizedStr, realizedStr, liqStr string
		if err := rows.Scan(&traderIDStr, &pos.Instrument, &sizeStr, &entryStr, &pos.Leverage, &marginStr, &unrealizedStr, &realizedStr, &liqStr, &pos.UpdatedAt); err != nil {
			return nil, err
		}
		pos.TraderID, _ = uuid.Parse(traderIDStr)
		pos.Size, _ = decimal.NewFromString(sizeStr)
		pos.EntryPrice, _ = decimal.NewFromString(entryStr)
		pos.Margin, _ = decimal.NewFromString(marginStr)
		pos.UnrealizedPnL, _ = decimal.NewFromString(unrealizedStr)
		pos.RealizedPnL, _ = decimal.NewFromString(realizedStr)
		pos.LiquidationPrice, _ = decimal.NewFromString(liqStr)
		positions = append(positions, &pos)
	}

	return positions, nil
}

// === Order Operations ===

// SaveOrder inserts or updates an order
func (s *SQLiteDB) SaveOrder(order *domain.Order) error {
	query := `
	INSERT INTO orders (id, trader_id, instrument, side, type, price, size, filled_size, status, leverage, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		filled_size = excluded.filled_size,
		status = excluded.status,
		updated_at = excluded.updated_at
	`
	_, err := s.db.Exec(query,
		order.ID.String(),
		order.TraderID.String(),
		order.Instrument,
		string(order.Side),
		string(order.Type),
		order.Price.String(),
		order.Size.String(),
		order.FilledSize.String(),
		string(order.Status),
		order.Leverage,
		order.CreatedAt,
		order.UpdatedAt,
	)
	return err
}

// DeleteOrder removes an order
func (s *SQLiteDB) DeleteOrder(orderID uuid.UUID) error {
	_, err := s.db.Exec("DELETE FROM orders WHERE id = ?", orderID.String())
	return err
}

// GetOpenOrders retrieves open orders for an instrument
func (s *SQLiteDB) GetOpenOrders(instrument string) ([]*domain.Order, error) {
	query := `SELECT id, trader_id, instrument, side, type, price, size, filled_size, status, leverage, created_at, updated_at FROM orders WHERE instrument = ? AND status = 'open' ORDER BY created_at`
	rows, err := s.db.Query(query, instrument)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		var order domain.Order
		var idStr, traderIDStr, sideStr, typeStr, priceStr, sizeStr, filledStr, statusStr string
		if err := rows.Scan(&idStr, &traderIDStr, &order.Instrument, &sideStr, &typeStr, &priceStr, &sizeStr, &filledStr, &statusStr, &order.Leverage, &order.CreatedAt, &order.UpdatedAt); err != nil {
			return nil, err
		}
		order.ID, _ = uuid.Parse(idStr)
		order.TraderID, _ = uuid.Parse(traderIDStr)
		order.Side = domain.Side(sideStr)
		order.Type = domain.OrderType(typeStr)
		order.Price, _ = decimal.NewFromString(priceStr)
		order.Size, _ = decimal.NewFromString(sizeStr)
		order.FilledSize, _ = decimal.NewFromString(filledStr)
		order.Status = domain.OrderStatus(statusStr)
		orders = append(orders, &order)
	}

	return orders, nil
}

// === Trade Operations ===

// SaveTrade inserts a trade
func (s *SQLiteDB) SaveTrade(trade *domain.Trade) error {
	query := `
	INSERT INTO trades (id, instrument, price, size, buyer_id, seller_id, buyer_leverage, seller_leverage, buyer_effect, seller_effect, aggressor_side, timestamp)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query,
		trade.ID.String(),
		trade.Instrument,
		trade.Price.String(),
		trade.Size.String(),
		trade.BuyerID.String(),
		trade.SellerID.String(),
		trade.BuyerLeverage,
		trade.SellerLeverage,
		string(trade.BuyerEffect),
		string(trade.SellerEffect),
		string(trade.AggressorSide),
		trade.Timestamp,
	)
	return err
}

// GetRecentTrades retrieves recent trades for an instrument
func (s *SQLiteDB) GetRecentTrades(instrument string, limit int) ([]*domain.Trade, error) {
	query := `SELECT id, instrument, price, size, buyer_id, seller_id, buyer_leverage, seller_leverage, buyer_effect, seller_effect, aggressor_side, timestamp FROM trades WHERE instrument = ? ORDER BY timestamp DESC LIMIT ?`
	rows, err := s.db.Query(query, instrument, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trades []*domain.Trade
	for rows.Next() {
		var trade domain.Trade
		var idStr, buyerIDStr, sellerIDStr, priceStr, sizeStr, buyerEffectStr, sellerEffectStr, aggressorStr string
		if err := rows.Scan(&idStr, &trade.Instrument, &priceStr, &sizeStr, &buyerIDStr, &sellerIDStr, &trade.BuyerLeverage, &trade.SellerLeverage, &buyerEffectStr, &sellerEffectStr, &aggressorStr, &trade.Timestamp); err != nil {
			return nil, err
		}
		trade.ID, _ = uuid.Parse(idStr)
		trade.BuyerID, _ = uuid.Parse(buyerIDStr)
		trade.SellerID, _ = uuid.Parse(sellerIDStr)
		trade.Price, _ = decimal.NewFromString(priceStr)
		trade.Size, _ = decimal.NewFromString(sizeStr)
		trade.BuyerEffect = domain.PositionEffect(buyerEffectStr)
		trade.SellerEffect = domain.PositionEffect(sellerEffectStr)
		trade.AggressorSide = domain.Side(aggressorStr)
		trades = append(trades, &trade)
	}

	return trades, nil
}

// GetTraderTrades retrieves trades for a specific trader
func (s *SQLiteDB) GetTraderTrades(traderID uuid.UUID, instrument string, limit int) ([]*domain.Trade, error) {
	query := `SELECT id, instrument, price, size, buyer_id, seller_id, buyer_leverage, seller_leverage, buyer_effect, seller_effect, aggressor_side, timestamp FROM trades WHERE instrument = ? AND (buyer_id = ? OR seller_id = ?) ORDER BY timestamp DESC LIMIT ?`
	rows, err := s.db.Query(query, instrument, traderID.String(), traderID.String(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trades []*domain.Trade
	for rows.Next() {
		var trade domain.Trade
		var idStr, buyerIDStr, sellerIDStr, priceStr, sizeStr, buyerEffectStr, sellerEffectStr, aggressorStr string
		if err := rows.Scan(&idStr, &trade.Instrument, &priceStr, &sizeStr, &buyerIDStr, &sellerIDStr, &trade.BuyerLeverage, &trade.SellerLeverage, &buyerEffectStr, &sellerEffectStr, &aggressorStr, &trade.Timestamp); err != nil {
			return nil, err
		}
		trade.ID, _ = uuid.Parse(idStr)
		trade.BuyerID, _ = uuid.Parse(buyerIDStr)
		trade.SellerID, _ = uuid.Parse(sellerIDStr)
		trade.Price, _ = decimal.NewFromString(priceStr)
		trade.Size, _ = decimal.NewFromString(sizeStr)
		trade.BuyerEffect = domain.PositionEffect(buyerEffectStr)
		trade.SellerEffect = domain.PositionEffect(sellerEffectStr)
		trade.AggressorSide = domain.Side(aggressorStr)
		trades = append(trades, &trade)
	}

	return trades, nil
}

// === Liquidation Operations ===

// SaveLiquidation inserts a liquidation
func (s *SQLiteDB) SaveLiquidation(liq *domain.Liquidation) error {
	query := `
	INSERT INTO liquidations (id, trader_id, instrument, side, size, entry_price, liquidation_price, mark_price, leverage, loss, insurance_fund_hit, timestamp)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	insuranceFundHit := 0
	if liq.InsuranceFundHit {
		insuranceFundHit = 1
	}
	_, err := s.db.Exec(query,
		liq.ID.String(),
		liq.TraderID.String(),
		liq.Instrument,
		string(liq.Side),
		liq.Size.String(),
		liq.EntryPrice.String(),
		liq.LiquidationPrice.String(),
		liq.MarkPrice.String(),
		liq.Leverage,
		liq.Loss.String(),
		insuranceFundHit,
		liq.Timestamp,
	)
	return err
}

// GetRecentLiquidations retrieves recent liquidations
func (s *SQLiteDB) GetRecentLiquidations(instrument string, limit int) ([]*domain.Liquidation, error) {
	query := `SELECT id, trader_id, instrument, side, size, entry_price, liquidation_price, mark_price, leverage, loss, insurance_fund_hit, timestamp FROM liquidations WHERE instrument = ? ORDER BY timestamp DESC LIMIT ?`
	rows, err := s.db.Query(query, instrument, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var liquidations []*domain.Liquidation
	for rows.Next() {
		var liq domain.Liquidation
		var idStr, traderIDStr, sideStr, sizeStr, entryStr, liqPriceStr, markStr, lossStr string
		var insuranceFundHit int
		if err := rows.Scan(&idStr, &traderIDStr, &liq.Instrument, &sideStr, &sizeStr, &entryStr, &liqPriceStr, &markStr, &liq.Leverage, &lossStr, &insuranceFundHit, &liq.Timestamp); err != nil {
			return nil, err
		}
		liq.ID, _ = uuid.Parse(idStr)
		liq.TraderID, _ = uuid.Parse(traderIDStr)
		liq.Side = domain.Side(sideStr)
		liq.Size, _ = decimal.NewFromString(sizeStr)
		liq.EntryPrice, _ = decimal.NewFromString(entryStr)
		liq.LiquidationPrice, _ = decimal.NewFromString(liqPriceStr)
		liq.MarkPrice, _ = decimal.NewFromString(markStr)
		liq.Loss, _ = decimal.NewFromString(lossStr)
		liq.InsuranceFundHit = insuranceFundHit == 1
		liquidations = append(liquidations, &liq)
	}

	return liquidations, nil
}

// === Market Stats Operations ===

// SaveMarketStats saves market statistics
func (s *SQLiteDB) SaveMarketStats(stats *domain.MarketStats) error {
	query := `
	INSERT INTO market_stats (instrument, last_price, mark_price, high_24h, low_24h, volume_24h, open_interest, insurance_fund, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(instrument) DO UPDATE SET
		last_price = excluded.last_price,
		mark_price = excluded.mark_price,
		high_24h = excluded.high_24h,
		low_24h = excluded.low_24h,
		volume_24h = excluded.volume_24h,
		open_interest = excluded.open_interest,
		insurance_fund = excluded.insurance_fund,
		updated_at = excluded.updated_at
	`
	_, err := s.db.Exec(query,
		stats.Instrument,
		stats.LastPrice.String(),
		stats.MarkPrice.String(),
		stats.High24h.String(),
		stats.Low24h.String(),
		stats.Volume24h.String(),
		stats.OpenInterest.String(),
		stats.InsuranceFund.String(),
		time.Now(),
	)
	return err
}

// GetMarketStats retrieves market statistics
func (s *SQLiteDB) GetMarketStats(instrument string) (*domain.MarketStats, error) {
	query := `SELECT instrument, last_price, mark_price, high_24h, low_24h, volume_24h, open_interest, insurance_fund FROM market_stats WHERE instrument = ?`
	row := s.db.QueryRow(query, instrument)

	var stats domain.MarketStats
	var lastPriceStr, markPriceStr, highStr, lowStr, volStr, oiStr, insuranceStr string
	err := row.Scan(&stats.Instrument, &lastPriceStr, &markPriceStr, &highStr, &lowStr, &volStr, &oiStr, &insuranceStr)
	if err == sql.ErrNoRows {
		// Return default stats
		return &domain.MarketStats{
			Instrument:    instrument,
			LastPrice:     decimal.NewFromInt(1000),
			MarkPrice:     decimal.NewFromInt(1000),
			High24h:       decimal.Zero,
			Low24h:        decimal.Zero,
			Volume24h:     decimal.Zero,
			OpenInterest:  decimal.Zero,
			InsuranceFund: decimal.NewFromInt(1000000),
		}, nil
	}
	if err != nil {
		return nil, err
	}

	stats.LastPrice, _ = decimal.NewFromString(lastPriceStr)
	stats.MarkPrice, _ = decimal.NewFromString(markPriceStr)
	stats.High24h, _ = decimal.NewFromString(highStr)
	stats.Low24h, _ = decimal.NewFromString(lowStr)
	stats.Volume24h, _ = decimal.NewFromString(volStr)
	stats.OpenInterest, _ = decimal.NewFromString(oiStr)
	stats.InsuranceFund, _ = decimal.NewFromString(insuranceStr)

	return &stats, nil
}
