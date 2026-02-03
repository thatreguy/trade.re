package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/thatreguy/trade.re/internal/domain"
)

// InsertTrade records a new trade
func (db *DB) InsertTrade(ctx context.Context, trade *domain.Trade) error {
	query := `
		INSERT INTO trades (id, instrument, price, size, buyer_id, seller_id, buyer_order_id, seller_order_id,
			buyer_leverage, seller_leverage, buyer_effect, seller_effect, buyer_new_position, seller_new_position, aggressor_side, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`
	_, err := db.Pool.Exec(ctx, query,
		trade.ID, trade.Instrument, trade.Price, trade.Size,
		trade.BuyerID, trade.SellerID, trade.BuyerOrderID, trade.SellerOrderID,
		trade.BuyerLeverage, trade.SellerLeverage,
		trade.BuyerEffect, trade.SellerEffect,
		trade.BuyerNewPosition, trade.SellerNewPosition,
		trade.AggressorSide, trade.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("inserting trade: %w", err)
	}
	return nil
}

// GetRecentTrades retrieves the most recent trades for an instrument
func (db *DB) GetRecentTrades(ctx context.Context, instrument string, limit int) ([]*domain.Trade, error) {
	query := `
		SELECT id, instrument, price, size, buyer_id, seller_id, buyer_order_id, seller_order_id,
			buyer_leverage, seller_leverage, buyer_effect, seller_effect, buyer_new_position, seller_new_position,
			aggressor_side, timestamp
		FROM trades WHERE instrument = $1
		ORDER BY timestamp DESC LIMIT $2
	`
	rows, err := db.Pool.Query(ctx, query, instrument, limit)
	if err != nil {
		return nil, fmt.Errorf("querying trades: %w", err)
	}
	defer rows.Close()

	var trades []*domain.Trade
	for rows.Next() {
		var t domain.Trade
		if err := rows.Scan(
			&t.ID, &t.Instrument, &t.Price, &t.Size,
			&t.BuyerID, &t.SellerID, &t.BuyerOrderID, &t.SellerOrderID,
			&t.BuyerLeverage, &t.SellerLeverage,
			&t.BuyerEffect, &t.SellerEffect,
			&t.BuyerNewPosition, &t.SellerNewPosition,
			&t.AggressorSide, &t.Timestamp,
		); err != nil {
			return nil, fmt.Errorf("scanning trade: %w", err)
		}
		trades = append(trades, &t)
	}
	return trades, nil
}

// GetTraderTrades retrieves trades for a specific trader
func (db *DB) GetTraderTrades(ctx context.Context, traderID uuid.UUID, limit int) ([]*domain.Trade, error) {
	query := `
		SELECT id, instrument, price, size, buyer_id, seller_id, buyer_order_id, seller_order_id,
			buyer_leverage, seller_leverage, buyer_effect, seller_effect, buyer_new_position, seller_new_position,
			aggressor_side, timestamp
		FROM trades WHERE buyer_id = $1 OR seller_id = $1
		ORDER BY timestamp DESC LIMIT $2
	`
	rows, err := db.Pool.Query(ctx, query, traderID, limit)
	if err != nil {
		return nil, fmt.Errorf("querying trader trades: %w", err)
	}
	defer rows.Close()

	var trades []*domain.Trade
	for rows.Next() {
		var t domain.Trade
		if err := rows.Scan(
			&t.ID, &t.Instrument, &t.Price, &t.Size,
			&t.BuyerID, &t.SellerID, &t.BuyerOrderID, &t.SellerOrderID,
			&t.BuyerLeverage, &t.SellerLeverage,
			&t.BuyerEffect, &t.SellerEffect,
			&t.BuyerNewPosition, &t.SellerNewPosition,
			&t.AggressorSide, &t.Timestamp,
		); err != nil {
			return nil, fmt.Errorf("scanning trade: %w", err)
		}
		trades = append(trades, &t)
	}
	return trades, nil
}

// InsertLiquidation records a liquidation event
func (db *DB) InsertLiquidation(ctx context.Context, liq *domain.Liquidation) error {
	query := `
		INSERT INTO liquidations (id, trader_id, instrument, side, size, entry_price, liquidation_price,
			mark_price, leverage, loss, counterparty_id, insurance_fund_hit, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	var counterpartyID interface{}
	if liq.CounterpartyID != uuid.Nil {
		counterpartyID = liq.CounterpartyID
	}

	_, err := db.Pool.Exec(ctx, query,
		liq.ID, liq.TraderID, liq.Instrument, liq.Side, liq.Size,
		liq.EntryPrice, liq.LiquidationPrice, liq.MarkPrice,
		liq.Leverage, liq.Loss, counterpartyID, liq.InsuranceFundHit, liq.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("inserting liquidation: %w", err)
	}
	return nil
}

// GetRecentLiquidations retrieves recent liquidation events
func (db *DB) GetRecentLiquidations(ctx context.Context, instrument string, limit int) ([]*domain.Liquidation, error) {
	query := `
		SELECT id, trader_id, instrument, side, size, entry_price, liquidation_price,
			mark_price, leverage, loss, counterparty_id, insurance_fund_hit, timestamp
		FROM liquidations WHERE instrument = $1
		ORDER BY timestamp DESC LIMIT $2
	`
	rows, err := db.Pool.Query(ctx, query, instrument, limit)
	if err != nil {
		return nil, fmt.Errorf("querying liquidations: %w", err)
	}
	defer rows.Close()

	var liquidations []*domain.Liquidation
	for rows.Next() {
		var l domain.Liquidation
		var counterpartyID *uuid.UUID
		if err := rows.Scan(
			&l.ID, &l.TraderID, &l.Instrument, &l.Side, &l.Size,
			&l.EntryPrice, &l.LiquidationPrice, &l.MarkPrice,
			&l.Leverage, &l.Loss, &counterpartyID, &l.InsuranceFundHit, &l.Timestamp,
		); err != nil {
			return nil, fmt.Errorf("scanning liquidation: %w", err)
		}
		if counterpartyID != nil {
			l.CounterpartyID = *counterpartyID
		}
		liquidations = append(liquidations, &l)
	}
	return liquidations, nil
}

// GetInsuranceFund retrieves the current insurance fund state
func (db *DB) GetInsuranceFund(ctx context.Context) (*domain.InsuranceFund, error) {
	query := `SELECT balance, total_in, total_out, updated_at FROM insurance_fund WHERE id = 1`
	var f domain.InsuranceFund
	err := db.Pool.QueryRow(ctx, query).Scan(&f.Balance, &f.TotalIn, &f.TotalOut, &f.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("querying insurance fund: %w", err)
	}
	return &f, nil
}

// UpdateInsuranceFund updates the insurance fund balance
func (db *DB) UpdateInsuranceFund(ctx context.Context, fund *domain.InsuranceFund) error {
	query := `UPDATE insurance_fund SET balance = $1, total_in = $2, total_out = $3 WHERE id = 1`
	_, err := db.Pool.Exec(ctx, query, fund.Balance, fund.TotalIn, fund.TotalOut)
	return err
}
