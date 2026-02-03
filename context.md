# Trade.re - Project Context

## Vision

Trade.re is a **fully transparent trading simulation game** where all market data, positions, and participant behavior are open for everyone to see. Unlike real markets where information asymmetry creates unfair advantages, Trade.re levels the playing field by making every piece of data public.

## Core Philosophy

### Radical Transparency
- Every trade is public: buyer, seller, price, size, timestamp
- Every position is visible: who holds what, at what entry price
- Every liquidation is tracked: no hidden forced closures
- Every market maker's inventory is exposed
- No dark pools, no hidden orders, no information advantage

### Open Interest Redefined
Traditional OI just shows aggregate numbers. Trade.re breaks it down:
- **OI Increase**: New long opened + New short opened (who matched with whom)
- **OI Decrease**: Long closed + Short closed (who exited against whom)
- **Liquidations**: Forced closures with full attribution (who got liquidated, who took the other side)

## Participant Types

### 1. Market Makers
- Provide liquidity by quoting bid/ask spreads
- Inventory and P&L fully visible
- Can be human or algorithmic
- Earn spread, take inventory risk

### 2. Real Traders (Humans)
- Manual trading via web interface
- All positions and history public
- Compete on strategy, not information

### 3. Bot Traders
- Algorithmic traders with open-source or disclosed strategies
- Must register and be identifiable
- Performance metrics public
- Creates a competitive algo arena

## Data Model

### Trade Record
```
{
  trade_id: uuid,
  timestamp: datetime,
  instrument: string,
  price: decimal,
  size: decimal,
  buyer_id: uuid,
  seller_id: uuid,
  buyer_position_effect: "open" | "close" | "liquidation",
  seller_position_effect: "open" | "close" | "liquidation",
  buyer_new_position: decimal,
  seller_new_position: decimal
}
```

### Position Record
```
{
  trader_id: uuid,
  instrument: string,
  size: decimal,  // positive = long, negative = short
  entry_price: decimal,
  unrealized_pnl: decimal,
  realized_pnl: decimal,
  liquidation_price: decimal,
  margin_used: decimal
}
```

### Open Interest Breakdown
```
{
  instrument: string,
  timestamp: datetime,
  total_oi: decimal,
  long_positions: count,
  short_positions: count,
  new_longs_opened: count,
  new_shorts_opened: count,
  longs_closed: count,
  shorts_closed: count,
  longs_liquidated: count,
  shorts_liquidated: count
}
```

## Technical Architecture

### Backend (Implemented)
- **Language**: Go 1.21+
  - Chosen for: excellent concurrency (goroutines), strong typing, single binary deployment
  - High performance for matching engine operations
- **Router**: chi/v5 (lightweight, idiomatic Go)
- **WebSocket**: gorilla/websocket for real-time feeds
- **Decimals**: shopspring/decimal for precise financial calculations
- **Database**: TimescaleDB (planned for persistence)

### Project Structure
```
trade.re/
├── cmd/server/          # Main entry point
├── internal/
│   ├── domain/          # Core types (Trade, Order, Position, etc.)
│   ├── engine/          # Matching engine & order book
│   ├── api/             # REST API handlers
│   └── ws/              # WebSocket hub & broadcasting
├── context.md           # This file
└── README.md
```

### Matching Engine (Implemented)
- **Order Book**: Price-level based with FIFO queuing at each level
- **Matching**: Price-time priority
- **Order Types**: Limit, Market
- **Self-trade Prevention**: Orders won't match against same trader
- **Position Tracking**: Automatic position effect detection (open/close)
- **P&L Calculation**: Real-time realized/unrealized P&L

### API Endpoints (Implemented)
```
GET  /health                              # Health check
GET  /ws                                  # WebSocket connection
GET  /api/v1/traders                      # List all traders (public)
POST /api/v1/traders                      # Register new trader
GET  /api/v1/traders/{id}                 # Get trader details
GET  /api/v1/traders/{id}/positions       # Get trader positions (public!)
GET  /api/v1/instruments/{symbol}/orderbook  # Order book snapshot
GET  /api/v1/instruments/{symbol}/positions  # ALL positions (transparency!)
GET  /api/v1/instruments/{symbol}/oi         # Open interest breakdown
POST /api/v1/orders                       # Submit order
DELETE /api/v1/orders/{id}                # Cancel order
```

### WebSocket Events
```json
{"type": "trade", "data": {...}}      // Every trade, with buyer/seller IDs
{"type": "order", "data": {...}}      // Order updates
{"type": "position", "data": {...}}   // Position changes
{"type": "orderbook", "data": {...}}  // Book updates (subscribe per instrument)
```

### Frontend (Planned)
- **Framework**: React/Next.js or SvelteKit
- **Charts**: TradingView lightweight charts or custom D3.js
- **Real-time**: WebSocket subscriptions

## Additional Ideas

### Gamification
- **Leaderboards**: Daily, weekly, all-time rankings by P&L, Sharpe ratio, win rate
- **Achievements**: Badges for milestones (first profitable trade, survived a crash, etc.)
- **Seasons**: Periodic resets with prizes for top performers
- **Tournaments**: Scheduled competitions with specific rules

### Educational Features
- **Replay Mode**: Watch historical market events unfold in real-time
- **Strategy Backtesting**: Test ideas against historical data
- **Tutorial Challenges**: Learn trading concepts through guided scenarios
- **Paper Trading**: Risk-free practice mode

### Social Features
- **Trader Profiles**: Public track record, strategy descriptions
- **Follow System**: Get notifications when followed traders make moves
- **Copy Trading**: Automatically mirror successful traders (with consent)
- **Chat/Forums**: Discuss strategies and market conditions

### Analytics Dashboard
- **Whale Tracking**: Alerts when large positions open/close
- **Sentiment Indicators**: Long/short ratio, funding rate predictions
- **Heat Maps**: Liquidation clusters, support/resistance levels
- **Flow Analysis**: Net buying/selling by trader type

### Market Mechanics
- **Funding Rates**: Periodic payments between longs and shorts
- **Insurance Fund**: Pool to cover liquidation shortfalls
- **Circuit Breakers**: Automatic halts during extreme volatility
- **Multiple Instruments**: Perpetuals, dated futures, options

### Bot Ecosystem
- **SDK/API**: Easy-to-use libraries for building bots
- **Strategy Marketplace**: Share/sell trading algorithms
- **Sandboxed Execution**: Run bots in controlled environment
- **Performance Attribution**: Detailed bot analytics

### Transparency Reports
- **Daily Summaries**: Aggregate statistics and notable events
- **Market Health Metrics**: Liquidity depth, spread analysis
- **Anomaly Detection**: Flag unusual trading patterns
- **Open Data Export**: Full historical data downloads

## Success Metrics

1. **Adoption**: Number of active traders (human + bot)
2. **Liquidity**: Average bid-ask spread, order book depth
3. **Engagement**: Trades per user, session duration
4. **Education**: Tutorial completion rates, strategy diversity
5. **Community**: Forum activity, shared strategies

## Roadmap Phases

### Phase 1: Foundation ✅ IN PROGRESS
- [x] Core matching engine (Go)
- [x] Order book with price-time priority
- [x] REST API for trading operations
- [x] WebSocket real-time feeds
- [x] Position tracking with P&L
- [x] Two perpetual instruments (BTC-PERP, ETH-PERP)
- [ ] Database persistence (TimescaleDB)
- [ ] Basic web interface
- [ ] Authentication system

### Phase 2: Transparency
- [x] OI breakdown endpoint
- [x] Position explorer (all positions public)
- [ ] Trade history search with filters
- [ ] Historical data API
- [ ] Liquidation tracking

### Phase 3: Social
- [ ] Trader profiles with track record
- [ ] Leaderboards (P&L, Sharpe, win rate)
- [ ] Follow system
- [ ] Activity feed

### Phase 4: Ecosystem
- [ ] Bot SDK (Go, Python, TypeScript)
- [ ] Strategy marketplace
- [ ] Tournaments
- [ ] Mobile app

## Development

### Running the Server
```bash
# Build and run
go run ./cmd/server

# Or build binary
go build -o trade.re ./cmd/server
./trade.re

# With custom port
PORT=3000 go run ./cmd/server
```

### Testing the API
```bash
# Create a trader
curl -X POST http://localhost:8080/api/v1/traders \
  -H "Content-Type: application/json" \
  -d '{"username": "alice", "type": "human"}'

# Submit a limit order
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "trader_id": "<trader-uuid>",
    "instrument": "BTC-PERP",
    "side": "buy",
    "type": "limit",
    "price": "50000",
    "size": "1"
  }'

# Get order book
curl http://localhost:8080/api/v1/instruments/BTC-PERP/orderbook

# Get all positions (transparency!)
curl http://localhost:8080/api/v1/instruments/BTC-PERP/positions

# Get OI breakdown
curl http://localhost:8080/api/v1/instruments/BTC-PERP/oi
```

### WebSocket Testing
```javascript
const ws = new WebSocket('ws://localhost:8080/ws');
ws.onmessage = (e) => console.log(JSON.parse(e.data));
// Subscribe to order book updates
ws.send(JSON.stringify({type: 'subscribe', data: 'orderbook:BTC-PERP'}));
```

## Open Questions

1. Should there be any virtual currency/points, or use play money?
2. How to prevent wash trading and self-dealing?
3. What anti-manipulation rules should exist?
4. How to balance bot vs human traders fairly?
5. Should market maker positions have different visibility rules?

## References

- [BitMEX Research on OI](https://blog.bitmex.com/)
- [Deribit Insights](https://insights.deribit.com/)
- [Paradigm Research](https://www.paradigm.xyz/)
