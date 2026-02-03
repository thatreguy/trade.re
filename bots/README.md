# Trade.re Bots

Trading bots for the Trade.re platform. All bots use the REST API - no special SDK needed!

## Available Bots

### 1. Market Maker (`market_maker.py`)

Provides liquidity by posting bid/ask orders around the current price.

**Features:**
- Posts orders at multiple price levels
- Configurable spread and position size
- Uses conservative leverage (5x default)
- Automatically registers as `market_maker` type

**Usage:**
```bash
pip install -r requirements.txt
python market_maker.py
```

**Configuration:**
- `spread`: Bid/ask spread percentage (default: 0.5%)
- `order_size`: Size per order (default: 5.0)
- `leverage`: Leverage to use (default: 5x)
- `num_levels`: Number of price levels (default: 3)

### 2. News Sentiment Trader (`news_trader.py`)

Curates world news and takes bullish/bearish positions based on sentiment.

**Features:**
- Analyzes news headlines for sentiment
- Takes long positions on bullish sentiment
- Takes short positions on bearish sentiment
- Uses moderate leverage (25x default)
- Registers as `bot` type

**Usage:**
```bash
pip install -r requirements.txt
python news_trader.py
```

**Configuration:**
- `position_size`: Size per trade (default: 2.0)
- `leverage`: Leverage to use (default: 25x)
- `sentiment_threshold`: Min sentiment score to trade (default: 0.3)

**Optional:** Set `NEWS_API_KEY` environment variable for real news from NewsAPI.org. Otherwise uses mock news data.

## Transparency Note

All bot positions, trades, and leverage are **PUBLIC** on Trade.re. This is a core feature - everyone can see:
- What positions bots hold
- What leverage they're using
- Their trade history
- Their P&L

This levels the playing field between humans and bots.

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `API_URL` | Backend API URL | `http://localhost:8080` |
| `NEWS_API_KEY` | NewsAPI.org API key | (uses mock data) |

## Creating Your Own Bot

Bots are just HTTP clients! Here's the basic flow:

1. **Register** your bot:
```python
resp = requests.post(f"{API_URL}/api/v1/auth/register", json={
    "username": "my_bot",
    "password": "secret",
    "type": "bot"  # or "market_maker"
})
trader_id = resp.json()["trader"]["id"]
token = resp.json()["token"]
```

2. **Get market data**:
```python
# Order book
resp = requests.get(f"{API_URL}/api/v1/market/orderbook")

# Recent trades
resp = requests.get(f"{API_URL}/api/v1/market/trades")

# All positions (transparency!)
resp = requests.get(f"{API_URL}/api/v1/market/positions")
```

3. **Place orders**:
```python
resp = requests.post(f"{API_URL}/api/v1/orders", json={
    "trader_id": trader_id,
    "instrument": "R.index",
    "side": "buy",  # or "sell"
    "type": "limit",  # or "market"
    "price": "1000",
    "size": "1",
    "leverage": 10
}, headers={"Authorization": f"Bearer {token}"})
```

4. **Check your position** (everyone can see it!):
```python
resp = requests.get(f"{API_URL}/api/v1/traders/{trader_id}/positions")
```
