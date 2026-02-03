package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"github.com/thatreguy/trade.re/internal/domain"
)

// CreateTrader inserts a new trader
func (db *DB) CreateTrader(ctx context.Context, trader *domain.Trader) error {
	query := `
		INSERT INTO traders (id, username, type, password_hash, api_key_hash, balance, total_pnl, trade_count, max_leverage_used)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := db.Pool.Exec(ctx, query,
		trader.ID,
		trader.Username,
		trader.Type,
		trader.PasswordHash,
		trader.APIKeyHash,
		trader.Balance,
		trader.TotalPnL,
		trader.TradeCount,
		trader.MaxLeverageUsed,
	)
	if err != nil {
		return fmt.Errorf("inserting trader: %w", err)
	}
	return nil
}

// GetTrader retrieves a trader by ID
func (db *DB) GetTrader(ctx context.Context, id uuid.UUID) (*domain.Trader, error) {
	query := `
		SELECT id, username, type, password_hash, api_key_hash, balance, total_pnl, trade_count, max_leverage_used, created_at
		FROM traders WHERE id = $1
	`
	var t domain.Trader
	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.Username, &t.Type, &t.PasswordHash, &t.APIKeyHash,
		&t.Balance, &t.TotalPnL, &t.TradeCount, &t.MaxLeverageUsed, &t.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying trader: %w", err)
	}
	return &t, nil
}

// GetTraderByUsername retrieves a trader by username
func (db *DB) GetTraderByUsername(ctx context.Context, username string) (*domain.Trader, error) {
	query := `
		SELECT id, username, type, password_hash, api_key_hash, balance, total_pnl, trade_count, max_leverage_used, created_at
		FROM traders WHERE username = $1
	`
	var t domain.Trader
	err := db.Pool.QueryRow(ctx, query, username).Scan(
		&t.ID, &t.Username, &t.Type, &t.PasswordHash, &t.APIKeyHash,
		&t.Balance, &t.TotalPnL, &t.TradeCount, &t.MaxLeverageUsed, &t.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying trader by username: %w", err)
	}
	return &t, nil
}

// GetTraderByAPIKey retrieves a trader by API key hash
func (db *DB) GetTraderByAPIKey(ctx context.Context, apiKeyHash string) (*domain.Trader, error) {
	query := `
		SELECT id, username, type, password_hash, api_key_hash, balance, total_pnl, trade_count, max_leverage_used, created_at
		FROM traders WHERE api_key_hash = $1
	`
	var t domain.Trader
	err := db.Pool.QueryRow(ctx, query, apiKeyHash).Scan(
		&t.ID, &t.Username, &t.Type, &t.PasswordHash, &t.APIKeyHash,
		&t.Balance, &t.TotalPnL, &t.TradeCount, &t.MaxLeverageUsed, &t.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying trader by API key: %w", err)
	}
	return &t, nil
}

// GetAllTraders retrieves all traders (public info only)
func (db *DB) GetAllTraders(ctx context.Context) ([]*domain.Trader, error) {
	query := `
		SELECT id, username, type, balance, total_pnl, trade_count, max_leverage_used, created_at
		FROM traders ORDER BY total_pnl DESC
	`
	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying traders: %w", err)
	}
	defer rows.Close()

	var traders []*domain.Trader
	for rows.Next() {
		var t domain.Trader
		if err := rows.Scan(&t.ID, &t.Username, &t.Type, &t.Balance, &t.TotalPnL,
			&t.TradeCount, &t.MaxLeverageUsed, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning trader: %w", err)
		}
		traders = append(traders, &t)
	}
	return traders, nil
}

// UpdateTraderBalance updates a trader's balance
func (db *DB) UpdateTraderBalance(ctx context.Context, id uuid.UUID, balance decimal.Decimal) error {
	query := `UPDATE traders SET balance = $1 WHERE id = $2`
	_, err := db.Pool.Exec(ctx, query, balance, id)
	return err
}

// UpdateTraderPnL updates a trader's P&L and stats
func (db *DB) UpdateTraderPnL(ctx context.Context, id uuid.UUID, pnl decimal.Decimal, tradeCount int64) error {
	query := `UPDATE traders SET total_pnl = total_pnl + $1, trade_count = trade_count + $2 WHERE id = $3`
	_, err := db.Pool.Exec(ctx, query, pnl, tradeCount, id)
	return err
}

// UpdateTraderMaxLeverage updates max leverage if higher
func (db *DB) UpdateTraderMaxLeverage(ctx context.Context, id uuid.UUID, leverage int) error {
	query := `UPDATE traders SET max_leverage_used = GREATEST(max_leverage_used, $1) WHERE id = $2`
	_, err := db.Pool.Exec(ctx, query, leverage, id)
	return err
}

// UpdateTraderAPIKey sets the API key hash
func (db *DB) UpdateTraderAPIKey(ctx context.Context, id uuid.UUID, apiKeyHash string) error {
	query := `UPDATE traders SET api_key_hash = $1 WHERE id = $2`
	_, err := db.Pool.Exec(ctx, query, apiKeyHash, id)
	return err
}
