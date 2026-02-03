# Trade.re - Project Context

## Vision

Trade.re is a **fully transparent trading game** where participants trade a single instrument - **R.index** - representing collective sentiment on the global state of affairs. All market data, positions, and participant behavior are open for everyone to see. Unlike real markets where information asymmetry creates unfair advantages, Trade.re levels the playing field by making every piece of data public.

## Core Philosophy

### Radical Transparency
- Every trade is public: buyer, seller, price, size, timestamp
- Every position is visible: who holds what, at what entry price, **at what leverage**
- Every liquidation is tracked: no hidden forced closures
- Every market maker's inventory is exposed
- **Leverage is public**: See who's trading 1x vs 150x
- No dark pools, no hidden orders, no information advantage

### Open Interest Redefined
Traditional OI just shows aggregate numbers. Trade.re breaks it down:
- **OI Increase**: New long opened + New short opened (who matched with whom)
- **OI Decrease**: Long closed + Short closed (who exited against whom)
- **Liquidations**: Forced closures with full attribution (who got liquidated, who took the other side)

## The R.index

Trade.re features a single tradeable instrument: **R.index** - a perpetual index representing the **collective sentiment on the global state of affairs**.

### What R.index Represents
R.index is a crowdsourced barometer of global sentiment. When traders go long, they're expressing optimism about the world. When they short, they're expressing pessimism. The price emerges from the aggregate beliefs of all participants about how things are going globally - economically, politically, socially.

- **Bullish on R.index** = "I think the world is doing well / will improve"
- **Bearish on R.index** = "I think things are bad / will get worse"

### Why a Single Index?
- **Simplicity**: One market, maximum liquidity concentration
- **Collective Intelligence**: Price emerges from aggregate participant sentiment
- **Pure Expression**: No external oracle - the crowd IS the signal
- **Configurability**: Starting price set via config file

### R.index Specifications
| Parameter | Value |
|-----------|-------|
| Symbol | `R.index` |
| Type | Perpetual (no expiry) |
| Starting Price | Configurable (default: 1000) |
| Tick Size | 0.01 |
| Min Order Size | 0.001 |
| Max Leverage | 150x |
| Market Hours | **24/7** (always open) |
| Daily Candle Start | 00:00 UTC (5:30 AM IST) |

### Market Hours
- **24/7 Trading**: The market never closes
- **Daily Reset**: Statistics reset at 00:00 UTC (5:30 AM IST)
- **Candle Alignment**: All daily candles start at 00:00 UTC

## Leverage System

### Public Leverage (Core Transparency Feature!)
Every trader's leverage is **publicly visible**:
- When you open a position, your leverage choice is broadcast
- Position explorer shows: trader, size, entry, **leverage**, liquidation price
- Leaderboards can filter by leverage tier (1-10x, 10-50x, 50-150x)

### Leverage Tiers
| Tier | Range | Maintenance Margin | Color |
|------|-------|-------------------|-------|
| Conservative | 1-10x | 0.5% | Green |
| Moderate | 11-50x | 1% | Yellow |
| Aggressive | 51-100x | 2% | Red |
| Degen | 101-150x | 5% | Purple |

### Liquidation Rules
- **Liquidation Price** = Entry ± (Entry / Leverage) × (1 - Maintenance Margin)
- Liquidations are executed at mark price
- Insurance fund absorbs losses exceeding margin
- All liquidations broadcast in real-time with full details

## Participant Types

### 1. Market Makers
- Provide liquidity by quoting bid/ask spreads
- Inventory, P&L, and **leverage** fully visible
- Can be human or bot
- Earn spread, take inventory risk

### 2. Human Traders
- Manual trading via web interface
- All positions, history, and **leverage choices** public
- Compete on strategy, not information

### 3. Bot Traders
- Algorithmic traders using REST API
- Must register with `type: bot`
- Performance metrics and **leverage usage** public
- No special SDK needed - just use the REST API

## Data Model

### Trader Record
```json
{
  "id": "uuid",
  "username": "string",
  "type": "human | bot | market_maker",
  "balance": "decimal",
  "total_pnl": "decimal",
  "trade_count": "int",
  "max_leverage_used": "int"
}
```

### Position Record (Public!)
```json
{
  "trader_id": "uuid",
  "instrument": "R.index",
  "size": "decimal",
  "entry_price": "decimal",
  "leverage": "int",
  "margin": "decimal",
  "unrealized_pnl": "decimal",
  "liquidation_price": "decimal"
}
```

### Trade Record
```json
{
  "id": "uuid",
  "timestamp": "datetime",
  "price": "decimal",
  "size": "decimal",
  "buyer_id": "uuid",
  "seller_id": "uuid",
  "buyer_leverage": "int",
  "seller_leverage": "int",
  "buyer_effect": "open | close | liquidation",
  "seller_effect": "open | close | liquidation"
}
```

### Liquidation Record
```json
{
  "id": "uuid",
  "trader_id": "uuid",
  "side": "long | short",
  "size": "decimal",
  "entry_price": "decimal",
  "liquidation_price": "decimal",
  "mark_price": "decimal",
  "leverage": "int",
  "loss": "decimal",
  "insurance_fund_hit": "boolean",
  "timestamp": "datetime"
}
```

## Technical Architecture

### Backend (Go)
- **Language**: Go 1.21+
- **Router**: chi/v5
- **WebSocket**: gorilla/websocket
- **Decimals**: shopspring/decimal
- **Config**: YAML
- **Auth**: JWT + API keys

### Database (PostgreSQL)
- `traders` - User accounts and stats
- `positions` - Current open positions
- `orders` - Active and historical orders
- `trades` - Complete trade history
- `liquidations` - Liquidation events
- `insurance_fund` - Fund balance

### Project Structure
```
trade.re/
├── cmd/server/              # Main entry point
├── config/
│   └── config.yaml          # Server configuration
├── internal/
│   ├── config/              # Config loading
│   ├── domain/              # Core types
│   ├── engine/              # Matching engine & order book
│   ├── liquidation/         # Liquidation engine
│   ├── api/                 # REST API handlers
│   ├── auth/                # Authentication
│   ├── db/                  # Database layer
│   └── ws/                  # WebSocket hub
├── web/                     # Frontend (Next.js)
├── schema.sql               # Database schema
├── context.md               # This file
└── README.md
```

### Configuration (config.yaml)
```yaml
server:
  port: 8080
  host: "0.0.0.0"

database:
  host: localhost
  port: 5432
  name: tradere
  user: tradere
  password: ${DB_PASSWORD}

rindex:
  starting_price: 1000
  tick_size: 0.01
  min_order_size: 0.001
  max_leverage: 150

auth:
  jwt_secret: ${JWT_SECRET}
  token_expiry_hours: 24

liquidation:
  check_interval_ms: 100
  insurance_fund_initial: 1000000

game:
  starting_balance: 10000
```

### API Endpoints
```
# Health & Info
GET  /health
GET  /api/v1/config                        # Public config

# Auth
POST /api/v1/auth/register                 # Register trader
POST /api/v1/auth/login                    # Get JWT token
POST /api/v1/auth/api-key                  # Generate API key

# Traders (Public!)
GET  /api/v1/traders                       # All traders
GET  /api/v1/traders/{id}                  # Trader details
GET  /api/v1/traders/{id}/positions        # Trader positions
GET  /api/v1/traders/{id}/trades           # Trade history

# Market (Public!)
GET  /api/v1/market/orderbook              # Order book
GET  /api/v1/market/positions              # ALL positions
GET  /api/v1/market/oi                     # Open interest breakdown
GET  /api/v1/market/trades                 # Recent trades
GET  /api/v1/market/liquidations           # Recent liquidations
GET  /api/v1/market/stats                  # Market statistics
GET  /api/v1/market/candles                # OHLCV candles (1m, 5m, 1h, 1d)

# Trading (Authenticated)
POST   /api/v1/orders                      # Submit order
DELETE /api/v1/orders/{id}                 # Cancel order
POST   /api/v1/positions/close             # Close position

# WebSocket
GET /ws                                    # Real-time feed
```

### WebSocket Events
```json
{"type": "trade", "data": {...}}           // Every trade
{"type": "order", "data": {...}}           // Order updates
{"type": "position", "data": {...}}        // Position changes
{"type": "liquidation", "data": {...}}     // Liquidations
{"type": "orderbook", "data": {...}}       // Book updates
```

## Liquidation Engine

### How It Works
1. **Continuous Monitoring**: Check positions every 100ms
2. **Mark Price**: Order book mid-price
3. **Trigger**: When mark crosses liquidation price
4. **Execution**: Close at market, insurance fund absorbs excess loss

### Insurance Fund
- Seeded with configurable initial amount (default: 1M)
- Grows from liquidation profits (margin > loss)
- Depletes when loss > margin
- Balance is public

## Web Frontend

### Tech Stack
- **Framework**: Next.js 14
- **Styling**: Tailwind CSS
- **Charts**: TradingView Lightweight Charts
- **State**: Zustand
- **Real-time**: WebSocket

### Pages
1. **Trading** (`/`) - Chart, order book, trade form, positions
2. **All Positions** (`/positions`) - Transparency dashboard
3. **Leaderboard** (`/leaderboard`) - P&L rankings by leverage tier
4. **Liquidations** (`/liquidations`) - Real-time liquidation feed
5. **Trader Profile** (`/trader/{id}`) - Public history and stats

### State Management (Zustand)
Two stores in `web/src/store/`:

**Market Store** (`market.ts`)
- **State**: `orderBook`, `recentTrades`, `allPositions`, `recentLiquidations`, `marketStats`, `traders`
- **Actions**: `fetchOrderBook()`, `fetchRecentTrades()`, `fetchAllPositions()`, `fetchAll()`, etc.
- **WebSocket**: `connectWebSocket()` subscribes to real-time updates (trades, orderbook, positions, liquidations)
- **Real-time handlers**: `addTrade()`, `updateOrderBook()`, `updatePosition()`, `addLiquidation()`

**User Store** (`user.ts`)
- **State**: `user`, `isAuthenticated`, `position`, `trades`
- **Actions**: `login()`, `register()`, `logout()`, `checkAuth()`, `fetchUserData()`

### API Client (`web/src/lib/`)
- `api.ts` - REST API client with typed responses
- `websocket.ts` - WebSocket client for real-time updates

## Authentication

### Flow
1. **Register**: Username + password → account created
2. **Login**: Credentials → JWT token (24h expiry)
3. **API Key**: Generate long-lived key for bots

### Public vs Private
| Data | Visibility |
|------|------------|
| Positions | Public |
| Leverage | Public |
| Trade history | Public |
| P&L | Public |
| Order book | Public |
| Liquidations | Public |
| Password | Private |
| API keys | Private |

## Roadmap

### Phase 1: Core ✅ DONE
- [x] Matching engine
- [x] Order book with FIFO
- [x] REST API
- [x] WebSocket feeds
- [x] Domain types with leverage
- [x] PostgreSQL schema
- [x] Config system
- [x] Auth (JWT + API keys)
- [x] Liquidation engine
- [x] Basic frontend (trading, positions)
- [x] Leaderboard page
- [x] Liquidations page
- [x] Trader profile page

### Phase 2: Integration (Current)
- [ ] Wire up API to frontend
- [ ] Real-time WebSocket in UI
- [ ] TradingView chart integration
- [ ] Candle data API (daily @ 00:00 UTC)

### Phase 3: Polish
- [ ] Historical data API
- [ ] Trade history search
- [ ] Position history
- [ ] Mobile responsive

### Phase 4: Social
- [ ] Follow traders
- [ ] Activity feed
- [ ] Tournaments

## Development

### Prerequisites
- Go 1.21+
- PostgreSQL 15+
- Node.js 18+

### Quick Start
```bash
# Clone
git clone https://github.com/thatreguy/trade.re.git
cd trade.re

# Database
createdb tradere
psql tradere < schema.sql

# Backend
go run ./cmd/server

# Frontend
cd web && npm install && npm run dev
```

### Bot Trading (REST API)
Bots don't need a special SDK - just use the REST API:

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username": "my_bot", "password": "secret", "type": "bot"}'

# Login and get token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "my_bot", "password": "secret"}' | jq -r '.token')

# Place order
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"side": "buy", "type": "limit", "price": "1000", "size": "1", "leverage": 10}'

# See ALL positions (transparency!)
curl http://localhost:8080/api/v1/market/positions
```

## Design Decisions

1. **No Funding Rate**: Keeps the game simpler. Price emerges purely from participant sentiment.
2. **Single Instrument (R.index)**: One index representing global sentiment - maximum liquidity, clear meaning.
3. **Public Leverage**: Core differentiator - see who's taking risk on their worldview.
4. **REST for Bots**: No SDK complexity - standard HTTP works everywhere.
5. **PostgreSQL**: Battle-tested, ACID compliant, great for financial data.
6. **24/7 Market**: Always open, no weekends. Daily candles align to 00:00 UTC.
7. **UTC for Everything**: All timestamps in UTC. Daily stats reset at midnight UTC (5:30 AM IST).
