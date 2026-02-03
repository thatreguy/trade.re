#!/usr/bin/env python3
"""
Random Trader Bot for Trade.re

This bot places random market orders to generate trading activity.
It alternates between buying and selling to create price movement.
All positions are PUBLIC - transparency is key!

Usage:
    python random_trader.py

Environment variables:
    API_URL: Backend API URL (default: http://localhost:8080)
"""

import os
import time
import random
import requests
from dataclasses import dataclass
from datetime import datetime

API_URL = os.getenv("API_URL", "http://localhost:8080")


@dataclass
class RandomTrader:
    trader_id: str
    token: str
    min_size: float = 1.0
    max_size: float = 5.0
    leverage: int = 10

    def get_orderbook(self) -> dict:
        """Get current orderbook"""
        resp = requests.get(f"{API_URL}/api/v1/market/orderbook")
        resp.raise_for_status()
        return resp.json()

    def place_order(self, side: str, order_type: str, price: float, size: float) -> dict:
        """Place an order"""
        resp = requests.post(
            f"{API_URL}/api/v1/orders",
            json={
                "trader_id": self.trader_id,
                "instrument": "R.index",
                "side": side,
                "type": order_type,
                "price": str(price),
                "size": str(size),
                "leverage": self.leverage
            },
            headers={"Authorization": f"Bearer {self.token}"}
        )
        resp.raise_for_status()
        return resp.json()

    def random_trade(self):
        """Execute a random trade"""
        try:
            orderbook = self.get_orderbook()
            bids = orderbook.get("bids", [])
            asks = orderbook.get("asks", [])

            if not bids and not asks:
                print(f"[RAND] No orders in book, skipping")
                return

            size = round(random.uniform(self.min_size, self.max_size), 2)

            # Randomly choose to buy or sell
            if random.random() > 0.5 and asks:
                # Buy (take ask)
                best_ask = float(asks[0]["price"])
                print(f"[RAND] {datetime.now().strftime('%H:%M:%S')} BUY {size} @ market (~${best_ask:.2f})")
                result = self.place_order("buy", "market", 0, size)
                trades = result.get("trades", [])
                if trades:
                    print(f"[RAND] Filled {len(trades)} trade(s)")
            elif bids:
                # Sell (take bid)
                best_bid = float(bids[0]["price"])
                print(f"[RAND] {datetime.now().strftime('%H:%M:%S')} SELL {size} @ market (~${best_bid:.2f})")
                result = self.place_order("sell", "market", 0, size)
                trades = result.get("trades", [])
                if trades:
                    print(f"[RAND] Filled {len(trades)} trade(s)")
            else:
                print(f"[RAND] No matching side available")

        except Exception as e:
            print(f"[RAND] Error: {e}")

    def run(self, interval: int = 5):
        """Run the random trader loop"""
        print(f"[RAND] Random Trader starting...")
        print(f"[RAND] Size range: {self.min_size}-{self.max_size}, Leverage: {self.leverage}x")
        print(f"[RAND] Trader ID: {self.trader_id}")
        print(f"[RAND] NOTE: All positions and leverage are PUBLIC")
        print()

        while True:
            self.random_trade()
            # Random delay between trades
            delay = interval + random.randint(-2, 3)
            time.sleep(max(2, delay))


def register_bot() -> tuple[str, str]:
    """Register a new trading bot"""
    username = f"rand_bot_{random.randint(1000, 9999)}"
    resp = requests.post(
        f"{API_URL}/api/v1/auth/register",
        json={
            "username": username,
            "password": "bot_password",
            "type": "bot"
        }
    )
    resp.raise_for_status()
    data = resp.json()
    print(f"[RAND] Registered as: {username}")
    return data["trader"]["id"], data["token"]


def main():
    print("=" * 50)
    print("  Trade.re Random Trading Bot")
    print("  Generating activity for R.index")
    print("=" * 50)
    print()

    # Register bot
    trader_id, token = register_bot()

    # Create and run random trader
    trader = RandomTrader(
        trader_id=trader_id,
        token=token,
        min_size=1.0,
        max_size=3.0,
        leverage=10
    )

    trader.run(interval=5)  # Trade roughly every 5 seconds


if __name__ == "__main__":
    main()
