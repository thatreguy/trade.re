-- Trade.re Database Schema
-- PostgreSQL 15+

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Traders table
CREATE TABLE traders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('human', 'bot', 'market_maker')),
    password_hash VARCHAR(255),
    api_key_hash VARCHAR(255),
    balance DECIMAL(20, 8) NOT NULL DEFAULT 10000,
    total_pnl DECIMAL(20, 8) NOT NULL DEFAULT 0,
    trade_count BIGINT NOT NULL DEFAULT 0,
    max_leverage_used INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_traders_username ON traders(username);
CREATE INDEX idx_traders_type ON traders(type);
CREATE INDEX idx_traders_api_key ON traders(api_key_hash) WHERE api_key_hash IS NOT NULL;

-- Positions table (current open positions)
CREATE TABLE positions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    trader_id UUID NOT NULL REFERENCES traders(id),
    instrument VARCHAR(20) NOT NULL DEFAULT 'R.index',
    size DECIMAL(20, 8) NOT NULL,
    entry_price DECIMAL(20, 8) NOT NULL,
    leverage INT NOT NULL CHECK (leverage >= 1 AND leverage <= 150),
    margin DECIMAL(20, 8) NOT NULL,
    unrealized_pnl DECIMAL(20, 8) NOT NULL DEFAULT 0,
    realized_pnl DECIMAL(20, 8) NOT NULL DEFAULT 0,
    liquidation_price DECIMAL(20, 8) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(trader_id, instrument)
);

CREATE INDEX idx_positions_trader ON positions(trader_id);
CREATE INDEX idx_positions_instrument ON positions(instrument);
CREATE INDEX idx_positions_leverage ON positions(leverage);
CREATE INDEX idx_positions_liquidation ON positions(liquidation_price);

-- Orders table
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    trader_id UUID NOT NULL REFERENCES traders(id),
    instrument VARCHAR(20) NOT NULL DEFAULT 'R.index',
    side VARCHAR(4) NOT NULL CHECK (side IN ('buy', 'sell')),
    type VARCHAR(10) NOT NULL CHECK (type IN ('limit', 'market')),
    price DECIMAL(20, 8),
    size DECIMAL(20, 8) NOT NULL,
    filled_size DECIMAL(20, 8) NOT NULL DEFAULT 0,
    leverage INT NOT NULL CHECK (leverage >= 1 AND leverage <= 150),
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'partial', 'filled', 'cancelled')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_orders_trader ON orders(trader_id);
CREATE INDEX idx_orders_instrument_status ON orders(instrument, status);
CREATE INDEX idx_orders_price ON orders(price) WHERE status IN ('pending', 'partial');

-- Trades table (complete history)
CREATE TABLE trades (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    instrument VARCHAR(20) NOT NULL DEFAULT 'R.index',
    price DECIMAL(20, 8) NOT NULL,
    size DECIMAL(20, 8) NOT NULL,
    buyer_id UUID NOT NULL REFERENCES traders(id),
    seller_id UUID NOT NULL REFERENCES traders(id),
    buyer_order_id UUID NOT NULL REFERENCES orders(id),
    seller_order_id UUID NOT NULL REFERENCES orders(id),
    buyer_leverage INT NOT NULL,
    seller_leverage INT NOT NULL,
    buyer_effect VARCHAR(20) NOT NULL CHECK (buyer_effect IN ('open', 'close', 'liquidation')),
    seller_effect VARCHAR(20) NOT NULL CHECK (seller_effect IN ('open', 'close', 'liquidation')),
    buyer_new_position DECIMAL(20, 8) NOT NULL,
    seller_new_position DECIMAL(20, 8) NOT NULL,
    aggressor_side VARCHAR(4) NOT NULL CHECK (aggressor_side IN ('buy', 'sell')),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_trades_instrument_time ON trades(instrument, timestamp DESC);
CREATE INDEX idx_trades_buyer ON trades(buyer_id, timestamp DESC);
CREATE INDEX idx_trades_seller ON trades(seller_id, timestamp DESC);
CREATE INDEX idx_trades_timestamp ON trades(timestamp DESC);

-- Liquidations table
CREATE TABLE liquidations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    trader_id UUID NOT NULL REFERENCES traders(id),
    instrument VARCHAR(20) NOT NULL DEFAULT 'R.index',
    side VARCHAR(4) NOT NULL CHECK (side IN ('buy', 'sell')),
    size DECIMAL(20, 8) NOT NULL,
    entry_price DECIMAL(20, 8) NOT NULL,
    liquidation_price DECIMAL(20, 8) NOT NULL,
    mark_price DECIMAL(20, 8) NOT NULL,
    leverage INT NOT NULL,
    loss DECIMAL(20, 8) NOT NULL,
    counterparty_id UUID REFERENCES traders(id),
    insurance_fund_hit BOOLEAN NOT NULL DEFAULT FALSE,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_liquidations_trader ON liquidations(trader_id, timestamp DESC);
CREATE INDEX idx_liquidations_instrument_time ON liquidations(instrument, timestamp DESC);
CREATE INDEX idx_liquidations_leverage ON liquidations(leverage);

-- Insurance fund table
CREATE TABLE insurance_fund (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1), -- Singleton
    balance DECIMAL(20, 8) NOT NULL DEFAULT 1000000,
    total_in DECIMAL(20, 8) NOT NULL DEFAULT 0,
    total_out DECIMAL(20, 8) NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Initialize insurance fund
INSERT INTO insurance_fund (balance) VALUES (1000000) ON CONFLICT DO NOTHING;

-- Market stats table (for historical tracking)
CREATE TABLE market_stats (
    id SERIAL PRIMARY KEY,
    instrument VARCHAR(20) NOT NULL DEFAULT 'R.index',
    last_price DECIMAL(20, 8) NOT NULL,
    mark_price DECIMAL(20, 8) NOT NULL,
    high_24h DECIMAL(20, 8) NOT NULL,
    low_24h DECIMAL(20, 8) NOT NULL,
    volume_24h DECIMAL(20, 8) NOT NULL,
    open_interest DECIMAL(20, 8) NOT NULL,
    funding_rate DECIMAL(20, 8) NOT NULL DEFAULT 0,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_market_stats_time ON market_stats(instrument, timestamp DESC);

-- Funding payments table
CREATE TABLE funding_payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    trader_id UUID NOT NULL REFERENCES traders(id),
    instrument VARCHAR(20) NOT NULL DEFAULT 'R.index',
    position_size DECIMAL(20, 8) NOT NULL,
    funding_rate DECIMAL(20, 8) NOT NULL,
    payment DECIMAL(20, 8) NOT NULL, -- Positive = received, Negative = paid
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_funding_trader ON funding_payments(trader_id, timestamp DESC);
CREATE INDEX idx_funding_time ON funding_payments(timestamp DESC);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers for updated_at
CREATE TRIGGER traders_updated_at
    BEFORE UPDATE ON traders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER positions_updated_at
    BEFORE UPDATE ON positions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER orders_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER insurance_fund_updated_at
    BEFORE UPDATE ON insurance_fund
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- View: Active positions with trader info (for transparency dashboard)
CREATE VIEW v_positions_public AS
SELECT
    p.id,
    p.trader_id,
    t.username,
    t.type as trader_type,
    p.instrument,
    p.size,
    p.entry_price,
    p.leverage,
    p.margin,
    p.unrealized_pnl,
    p.realized_pnl,
    p.liquidation_price,
    p.updated_at
FROM positions p
JOIN traders t ON p.trader_id = t.id
WHERE p.size != 0;

-- View: Recent trades with trader info
CREATE VIEW v_trades_public AS
SELECT
    tr.id,
    tr.instrument,
    tr.price,
    tr.size,
    tr.buyer_id,
    tb.username as buyer_username,
    tb.type as buyer_type,
    tr.seller_id,
    ts.username as seller_username,
    ts.type as seller_type,
    tr.buyer_leverage,
    tr.seller_leverage,
    tr.buyer_effect,
    tr.seller_effect,
    tr.aggressor_side,
    tr.timestamp
FROM trades tr
JOIN traders tb ON tr.buyer_id = tb.id
JOIN traders ts ON tr.seller_id = ts.id;

-- View: Leaderboard by P&L
CREATE VIEW v_leaderboard AS
SELECT
    id,
    username,
    type,
    balance,
    total_pnl,
    trade_count,
    max_leverage_used,
    created_at
FROM traders
ORDER BY total_pnl DESC;
