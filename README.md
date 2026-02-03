# Trade.re

> **The Radically Transparent Trading Game**

Trade.re is an open-source trading simulation where **all data is public**. Every trade, every position, every liquidation - visible to everyone. No hidden orders. No information asymmetry. Just pure strategy.

## Why Trade.re?

In real markets, institutions have access to data retail traders never see. Trade.re flips that model:

- **See who bought from whom** - Not just price and volume, but the actual counterparties
- **Track every position** - Know exactly who's long, who's short, and at what price
- **Watch liquidations happen** - See who got liquidated and who took the other side
- **Analyze open interest properly** - OI broken down by new opens, closes, and forced liquidations

## Open Interest, Actually Open

Traditional exchanges show you a single OI number. We show you:

| Event | Description |
|-------|-------------|
| `+1 OI` | New long matched with new short (who + who) |
| `-1 OI` | Long closed against short close (who + who) |
| `Liquidation Long` | Forced closure, who got liquidated, who bought |
| `Liquidation Short` | Forced closure, who got liquidated, who sold |

## Participants

### Market Makers
Provide liquidity, earn spread. Their inventory? Public. Their P&L? Public.

### Traders
Humans competing on strategy, not information access. Full position history visible.

### Bots
Algorithmic traders that must register and be identifiable. Open arena for algo competition.

## Features

- **Real-time Order Book** - Full depth, all orders visible
- **Position Explorer** - Search any trader's positions
- **Trade Feed** - Live stream of all trades with counterparty info
- **Liquidation Alerts** - Know when positions get force-closed
- **Leaderboards** - Compete on P&L, Sharpe ratio, win rate
- **Public API** - Build your own bots and tools
- **Historical Data** - Full trade history, downloadable

## Tech Stack

```
Backend:    Rust/Go (matching engine) + Node.js (API)
Database:   TimescaleDB (time-series) + Redis (real-time)
Frontend:   Next.js + TradingView Charts
Real-time:  WebSockets
```

## Quick Start

```bash
# Clone the repo
git clone https://github.com/thatreguy/trade.re.git
cd trade.re

# Install dependencies
pnpm install

# Start development
pnpm dev
```

## Project Structure

```
trade.re/
├── apps/
│   ├── web/          # Next.js frontend
│   ├── api/          # REST API server
│   └── engine/       # Matching engine
├── packages/
│   ├── sdk/          # Trading bot SDK
│   ├── types/        # Shared TypeScript types
│   └── ui/           # Component library
├── docs/             # Documentation
└── data/             # Sample/test data
```

## API Preview

```typescript
// Get trader's positions
GET /api/v1/traders/{trader_id}/positions

// Get detailed OI breakdown
GET /api/v1/instruments/{symbol}/oi
{
  "total_oi": 15420,
  "breakdown": {
    "new_longs_opened": 234,
    "new_shorts_opened": 234,
    "longs_closed": 156,
    "shorts_closed": 156,
    "longs_liquidated": 12,
    "shorts_liquidated": 8
  }
}

// Stream all trades
WS /api/v1/stream/trades
{
  "trade_id": "...",
  "price": 50000.00,
  "size": 1.5,
  "buyer": { "id": "...", "effect": "open" },
  "seller": { "id": "...", "effect": "close" }
}
```

## Roadmap

- [ ] Core matching engine
- [ ] Basic web trading interface
- [ ] Real-time data feeds
- [ ] OI breakdown API
- [ ] Position explorer
- [ ] Trader profiles & leaderboards
- [ ] Bot SDK
- [ ] Mobile app

## Contributing

We're building this in the open. Contributions welcome!

1. Fork the repo
2. Create a feature branch
3. Make your changes
4. Submit a PR

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## Philosophy

> "Information wants to be free"

Markets work better when everyone has access to the same data. Trade.re is an experiment in radical transparency - what happens when you remove all information asymmetry from a market?

## License

MIT - See [LICENSE](LICENSE)

---

**Trade.re** - Where every trade tells the whole story.
