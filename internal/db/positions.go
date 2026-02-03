package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/thatreguy/trade.re/internal/domain"
)

// UpsertPosition creates or updates a position
func (db *DB) UpsertPosition(ctx context.Context, pos *domain.Position) error {
	query := `
		INSERT INTO positions (id, trader_id, instrument, size, entry_price, leverage, margin, unrealized_pnl, realized_pnl, liquidation_price)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (trader_id, instrument) DO UPDATE SET
			size = EXCLUDED.size,
			entry_price = EXCLUDED.entry_price,
			leverage = EXCLUDED.leverage,
			margin = EXCLUDED.margin,
			unrealized_pnl = EXCLUDED.unrealized_pnl,
			realized_pnl = EXCLUDED.realized_pnl,
			liquidation_price = EXCLUDED.liquidation_price
	`
	_, err := db.Pool.Exec(ctx, query,
		uuid.New(), pos.TraderID, pos.Instrument, pos.Size, pos.EntryPrice,
		pos.Leverage, pos.Margin, pos.UnrealizedPnL, pos.RealizedPnL, pos.LiquidationPrice,
	)
	if err != nil {
		return fmt.Errorf("upserting position: %w", err)
	}
	return nil
}

// GetPosition retrieves a trader's position for an instrument
func (db *DB) GetPosition(ctx context.Context, traderID uuid.UUID, instrument string) (*domain.Position, error) {
	query := `
		SELECT trader_id, instrument, size, entry_price, leverage, margin, unrealized_pnl, realized_pnl, liquidation_price, updated_at
		FROM positions WHERE trader_id = $1 AND instrument = $2
	`
	var p domain.Position
	err := db.Pool.QueryRow(ctx, query, traderID, instrument).Scan(
		&p.TraderID, &p.Instrument, &p.Size, &p.EntryPrice, &p.Leverage,
		&p.Margin, &p.UnrealizedPnL, &p.RealizedPnL, &p.LiquidationPrice, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying position: %w", err)
	}
	return &p, nil
}

// GetAllPositions retrieves all non-zero positions for an instrument (TRANSPARENCY!)
func (db *DB) GetAllPositions(ctx context.Context, instrument string) ([]*domain.Position, error) {
	query := `
		SELECT trader_id, instrument, size, entry_price, leverage, margin, unrealized_pnl, realized_pnl, liquidation_price, updated_at
		FROM positions WHERE instrument = $1 AND size != 0
		ORDER BY ABS(size) DESC
	`
	rows, err := db.Pool.Query(ctx, query, instrument)
	if err != nil {
		return nil, fmt.Errorf("querying positions: %w", err)
	}
	defer rows.Close()

	var positions []*domain.Position
	for rows.Next() {
		var p domain.Position
		if err := rows.Scan(&p.TraderID, &p.Instrument, &p.Size, &p.EntryPrice, &p.Leverage,
			&p.Margin, &p.UnrealizedPnL, &p.RealizedPnL, &p.LiquidationPrice, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning position: %w", err)
		}
		positions = append(positions, &p)
	}
	return positions, nil
}

// GetTraderPositions retrieves all positions for a trader (TRANSPARENCY!)
func (db *DB) GetTraderPositions(ctx context.Context, traderID uuid.UUID) ([]*domain.Position, error) {
	query := `
		SELECT trader_id, instrument, size, entry_price, leverage, margin, unrealized_pnl, realized_pnl, liquidation_price, updated_at
		FROM positions WHERE trader_id = $1 AND size != 0
	`
	rows, err := db.Pool.Query(ctx, query, traderID)
	if err != nil {
		return nil, fmt.Errorf("querying trader positions: %w", err)
	}
	defer rows.Close()

	var positions []*domain.Position
	for rows.Next() {
		var p domain.Position
		if err := rows.Scan(&p.TraderID, &p.Instrument, &p.Size, &p.EntryPrice, &p.Leverage,
			&p.Margin, &p.UnrealizedPnL, &p.RealizedPnL, &p.LiquidationPrice, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning position: %w", err)
		}
		positions = append(positions, &p)
	}
	return positions, nil
}

// GetPositionsNearLiquidation finds positions close to liquidation price
func (db *DB) GetPositionsNearLiquidation(ctx context.Context, instrument string, currentPrice float64, threshold float64) ([]*domain.Position, error) {
	query := `
		SELECT trader_id, instrument, size, entry_price, leverage, margin, unrealized_pnl, realized_pnl, liquidation_price, updated_at
		FROM positions
		WHERE instrument = $1 AND size != 0
		AND ABS(liquidation_price - $2) / $2 < $3
		ORDER BY ABS(liquidation_price - $2) ASC
	`
	rows, err := db.Pool.Query(ctx, query, instrument, currentPrice, threshold)
	if err != nil {
		return nil, fmt.Errorf("querying positions near liquidation: %w", err)
	}
	defer rows.Close()

	var positions []*domain.Position
	for rows.Next() {
		var p domain.Position
		if err := rows.Scan(&p.TraderID, &p.Instrument, &p.Size, &p.EntryPrice, &p.Leverage,
			&p.Margin, &p.UnrealizedPnL, &p.RealizedPnL, &p.LiquidationPrice, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning position: %w", err)
		}
		positions = append(positions, &p)
	}
	return positions, nil
}

// DeletePosition removes a position (when closed)
func (db *DB) DeletePosition(ctx context.Context, traderID uuid.UUID, instrument string) error {
	query := `DELETE FROM positions WHERE trader_id = $1 AND instrument = $2`
	_, err := db.Pool.Exec(ctx, query, traderID, instrument)
	return err
}
