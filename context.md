# Trade.re - Project Context

## Vision

Trade.re is a **fully transparent trading simulation game** where all market data, positions, and participant behavior are open for everyone to see. Unlike real markets where information asymmetry creates unfair advantages, Trade.re levels the playing field by making every piece of data public.

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

Trade.re features a single tradeable instrument: **R.index** - a virtual perpetual index.

### Why a Single Virtual Index?
- **Simplicity**: One market, maximum liquidity concentration
- **Fairness**: No information edge from external price feeds
- **Pure Gameplay**: Price is determined entirely by participant actions
- **Configurability**: Starting price set via config file

### R.index Specifications
| Parameter | Value |
|-----------|-------|
| Symbol | `R.index` |
| Type | Perpetual |
| Starting Price | Configurable (default: 1000) |
| Tick Size | 0.01 |
| Min Order Size | 0.001 |
| Max Leverage | 150x (Binance-style) |
| Funding Interval | 8 hours |

## Leverage System

### Public Leverage (Core Transparency Feature!)
Every trader's leverage is **publicly visible**:
- When you open a position, your leverage choice is broadcast
- Position explorer shows: trader, size, entry, **leverage**, liquidation price
- Leaderboards can filter by leverage tier (1-10x, 10-50x, 50-150x)

### Leverage Tiers
| Tier | Range | Maintenance Margin |
|------|-------|-------------------|
| Conservative | 1-10x | 0.5% |
| Moderate | 11-50x | 1% |
| Aggressive | 51-100x | 2% |
| Degen | 101-150x | 5% |

### Liquidation Rules
- **Liquidation Price** = Entry Price ± (Entry Price / Leverage) * (1 - Maintenance Margin)
- Liquidations are executed against the insurance fund first
- If insurance fund insufficient, ADL (Auto-Deleveraging) kicks in
- All liquidations are broadcast in real-time with full details

## Participant Types

### 1. Market Makers
- Provide liquidity by quoting bid/ask spreads
- Inventory, P&L, and **leverage** fully visible
- Can be human or algorithmic
- Earn spread, take inventory risk

### 2. Real Traders (Humans)
- Manual trading via web interface
- All positions, history, and **leverage choices** public
- Compete on strategy, not information

### 3. Bot Traders
- Algorithmic traders with open-source or disclosed strategies
- Must register and be identifiable
- Performance metrics and **leverage usage** public
- Creates a competitive algo arena

## Data Model

### Trader Record
```json
{
  "id": "uuid",
  "username": "string",
  "type": "human | bot | market_maker",
  "api_key_hash": "string",
  "created_at": "datetime",
  "total_pnl": "decimal",
  "trade_count": "int",
  "max_leverage_used": "int",
  "current_leverage": "int"
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
  "realized_pnl": "decimal",
  "liquidation_price": "decimal",
  "updated_at": "datetime"
}
```

### Trade Record
```json
{
  "id": "uuid",
  "timestamp": "datetime",
  "instrument": "R.index",
  "price": "decimal",
  "size": "decimal",
  "buyer_id": "uuid",
  "seller_id": "uuid",
  "buyer_leverage": "int",
  "seller_leverage": "int",
  "buyer_effect": "open | close | liquidation",
  "seller_effect": "open | close | liquidation",
  "buyer_new_position": "decimal",
  "seller_new_position": "decimal"
}
```

### Open Interest Breakdown
```json
{
  "instrument": "R.index",
  "timestamp": "datetime",
  "total_oi": "decimal",
  "long_positions": "int",
  "short_positions": "int",
  "avg_long_leverage": "decimal",
  "avg_short_leverage": "decimal",
  "new_longs_opened": "int",
  "new_shorts_opened": "int",
  "longs_closed": "int",
  "shorts_closed": "int",
  "longs_liquidated": "int",
  "shorts_liquidated": "int"
}
```

## Technical Architecture

### Backend
- **Language**: Go 1.21+
  - Excellent concurrency (goroutines)
  - Strong typing, single binary deployment
  - High performance for matching engine
- **Router**: chi/v5
- **WebSocket**: gorilla/websocket
- **Decimals**: shopspring/decimal
- **Config**: YAML config file
- **Auth**: JWT tokens with API keys

### Database: PostgreSQL
**Why PostgreSQL?**
- Battle-tested for financial applications
- ACID compliance for trade integrity
- JSON support for flexible data
- Excellent Go drivers (pgx)
- Can add TimescaleDB extension later for time-series optimization
- Rich ecosystem (backups, replication, monitoring)

**Schema Design**:
- `traders` - User accounts and stats
- `positions` - Current open positions
- `orders` - Active and historical orders
- `trades` - Complete trade history
- `liquidations` - Liquidation events
- `funding_payments` - Funding rate history

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
├── sdk/
│   ├── go/                  # Go SDK
│   ├── python/              # Python SDK
│   └── typescript/          # TypeScript SDK
├── web/                     # Frontend (React/Next.js)
├── context.md
└── README.md
```

### Configuration File (config.yaml)
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
  funding_interval_hours: 8

auth:
  jwt_secret: ${JWT_SECRET}
  token_expiry_hours: 24

liquidation:
  check_interval_ms: 100
  insurance_fund_initial: 1000000
```

### API Endpoints
```
# Health & Info
GET  /health
GET  /api/v1/config                        # Public config (starting price, etc.)

# Auth
POST /api/v1/auth/register                 # Register new trader
POST /api/v1/auth/login                    # Get JWT token
POST /api/v1/auth/api-key                  # Generate API key (for bots)

# Traders (Public - Transparency!)
GET  /api/v1/traders                       # List all traders
GET  /api/v1/traders/{id}                  # Trader details + leverage stats
GET  /api/v1/traders/{id}/positions        # Trader positions with leverage
GET  /api/v1/traders/{id}/trades           # Trade history

# R.index Market
GET  /api/v1/market/orderbook              # Order book
GET  /api/v1/market/positions              # ALL positions (transparency!)
GET  /api/v1/market/oi                     # Open interest breakdown
GET  /api/v1/market/trades                 # Recent trades
GET  /api/v1/market/liquidations           # Recent liquidations
GET  /api/v1/market/funding                # Current funding rate

# Trading (Authenticated)
POST   /api/v1/orders                      # Submit order (includes leverage)
DELETE /api/v1/orders/{id}                 # Cancel order
PUT    /api/v1/positions/leverage          # Adjust position leverage
POST   /api/v1/positions/close             # Close position

# WebSocket
GET /ws                                    # Real-time feed
```

### WebSocket Events
```json
{"type": "trade", "data": {...}}           // Every trade with leverage info
{"type": "order", "data": {...}}           // Order updates
{"type": "position", "data": {...}}        // Position changes (with leverage)
{"type": "liquidation", "data": {...}}     // Liquidation events
{"type": "orderbook", "data": {...}}       // Book updates
{"type": "funding", "data": {...}}         // Funding rate updates
```

## Liquidation Engine

### How It Works
1. **Continuous Monitoring**: Check all positions every 100ms
2. **Mark Price**: Use order book mid-price as mark
3. **Liquidation Trigger**: When mark price crosses liquidation price
4. **Execution Order**:
   - Try to close at market price
   - Insurance fund absorbs losses
   - If insurance depleted, trigger ADL

### Insurance Fund
- Seeded with configurable initial amount
- Grows from liquidation profits
- Depletes when liquidation losses exceed position margin
- Balance is public (transparency!)

### Auto-Deleveraging (ADL)
When insurance fund is depleted:
1. Rank opposite-side traders by profit + leverage
2. Force-close highest-ranked positions to cover
3. Notify affected traders in real-time

## Bot SDK

### Supported Languages
- **Go**: Native, first-class support
- **Python**: For quant/ML strategies
- **TypeScript**: For web-based bots

### SDK Features
```python
# Python SDK Example
from tradere import Client

client = Client(api_key="your-api-key")

# Get market data
orderbook = client.get_orderbook()
positions = client.get_all_positions()  # See everyone's positions!

# Place order with leverage
order = client.place_order(
    side="buy",
    size=1.0,
    price=1000.0,
    leverage=50  # Public!
)

# Stream real-time data
async for trade in client.stream_trades():
    print(f"{trade.buyer_id} bought from {trade.seller_id} at {trade.price}")
    print(f"Buyer leverage: {trade.buyer_leverage}x")
```

## Web Frontend

### Tech Stack
- **Framework**: Next.js 14 (App Router)
- **Styling**: Tailwind CSS
- **Charts**: Lightweight Charts (TradingView)
- **State**: Zustand
- **Real-time**: Native WebSocket

### Key Views
1. **Trading View**: Chart, order book, order entry, positions
2. **Transparency Dashboard**: All positions, OI breakdown, leverage distribution
3. **Leaderboard**: Rankings by P&L, filterable by leverage tier
4. **Trader Profile**: Public history, stats, leverage usage
5. **Liquidation Feed**: Real-time liquidation stream

## Authentication

### Simple Auth Flow
1. **Registration**: Username + password → JWT token
2. **Login**: Credentials → JWT token
3. **API Key**: For bots, generate long-lived API key
4. **Token Refresh**: Auto-refresh before expiry

### What's Public vs Private
| Data | Visibility |
|------|------------|
| Positions | Public |
| Leverage | Public |
| Trade history | Public |
| P&L | Public |
| Order book | Public |
| Open orders | Public |
| Password | Private |
| API keys | Private |
| JWT tokens | Private |

## Roadmap

### Phase 1: Core (Current Sprint)
- [x] Matching engine
- [x] Order book
- [x] REST API skeleton
- [x] WebSocket feeds
- [ ] **R.index instrument** (replacing BTC/ETH)
- [ ] **Leverage system (1-150x)**
- [ ] **PostgreSQL persistence**
- [ ] **Config file support**
- [ ] **Simple auth (JWT)**
- [ ] **Liquidation engine**
- [ ] **Bot SDK (Go)**
- [ ] **Basic web frontend**

### Phase 2: Polish
- [ ] Python SDK
- [ ] TypeScript SDK
- [ ] Funding rate mechanism
- [ ] Insurance fund
- [ ] ADL system
- [ ] Historical data API

### Phase 3: Social
- [ ] Leaderboards
- [ ] Trader profiles
- [ ] Follow system
- [ ] Activity feed

### Phase 4: Advanced
- [ ] Tournaments
- [ ] Strategy marketplace
- [ ] Mobile app

## Development

### Prerequisites
- Go 1.21+
- PostgreSQL 15+
- Node.js 18+ (for frontend)

### Quick Start
```bash
# Clone
git clone https://github.com/thatreguy/trade.re.git
cd trade.re

# Setup database
createdb tradere
psql tradere < schema.sql

# Configure
cp config/config.example.yaml config/config.yaml
# Edit config.yaml with your settings

# Run server
go run ./cmd/server

# Run frontend (separate terminal)
cd web && npm install && npm run dev
```

### Environment Variables
```bash
DB_PASSWORD=your_db_password
JWT_SECRET=your_jwt_secret_min_32_chars
```

## Open Questions

1. ~~Should there be virtual currency?~~ → Yes, play money with configurable starting balance
2. How to prevent wash trading? → Same-trader order matching already prevented
3. What anti-manipulation rules? → Position limits, order rate limits
4. Bot vs human fairness? → Separate leaderboards, same rules
5. ~~Leverage visibility?~~ → **Yes, fully public**

## References

- [BitMEX Perpetual Contracts](https://www.bitmex.com/app/perpetualContractsGuide)
- [Binance Futures Leverage](https://www.binance.com/en/support/faq/leverage-and-margin-of-usd%E2%93%A2-m-futures)
- [Deribit Liquidation](https://www.deribit.com/kb/liquidations)
