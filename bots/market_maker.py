#!/usr/bin/env python3
"""
Market Maker Bot for Trade.re

This bot provides liquidity by posting bid/ask orders around the current price.
All positions and leverage are PUBLIC - this is Trade.re's core transparency feature.

Usage:
    python market_maker.py

Environment variables:
    API_URL: Backend API URL (default: http://localhost:8080)
"""

import os
import time
import json
import random
import requests
from typing import Optional
from dataclasses import dataclass

API_URL = os.getenv("API_URL", "http://localhost:8080")

@dataclass
class MarketMaker:
    trader_id: str
    token: str
    spread: float = 0.5  # Spread percentage
    order_size: float = 5.0  # Size per order
    leverage: int = 5  # Conservative leverage
    num_levels: int = 3  # Number of price levels on each side

    def get_market_stats(self) -> dict:
        """Get current market stats"""
        resp = requests.get(f"{API_URL}/api/v1/market/stats")
        resp.raise_for_status()
        return resp.json()

    def get_orderbook(self) -> dict:
        """Get current orderbook"""
        resp = requests.get(f"{API_URL}/api/v1/market/orderbook")
        resp.raise_for_status()
        return resp.json()

    def place_order(self, side: str, price: float, size: float) -> dict:
        """Place a limit order"""
        resp = requests.post(
            f"{API_URL}/api/v1/orders",
            json={
                "trader_id": self.trader_id,
                "instrument": "R.index",
                "side": side,
                "type": "limit",
                "price": str(price),
                "size": str(size),
                "leverage": self.leverage
            },
            headers={"Authorization": f"Bearer {self.token}"}
        )
        resp.raise_for_status()
        return resp.json()

    def cancel_all_orders(self):
        """Cancel all open orders (simplified - just let them expire)"""
        pass  # In a real implementation, track and cancel orders

    def update_quotes(self):
        """Update bid/ask quotes around the current price"""
        try:
            stats = self.get_market_stats()
            mid_price = float(stats.get("last_price", 1000))

            # Calculate spread
            half_spread = mid_price * (self.spread / 100) / 2

            # Post orders at multiple levels
            for i in range(self.num_levels):
                level_offset = half_spread * (i + 1)

                # Bid (buy) orders
                bid_price = round(mid_price - level_offset, 2)
                try:
                    self.place_order("buy", bid_price, self.order_size)
                    print(f"[MM] Posted BID: {self.order_size} @ ${bid_price:.2f}")
                except Exception as e:
                    print(f"[MM] Failed to post bid: {e}")

                # Ask (sell) orders
                ask_price = round(mid_price + level_offset, 2)
                try:
                    self.place_order("sell", ask_price, self.order_size)
                    print(f"[MM] Posted ASK: {self.order_size} @ ${ask_price:.2f}")
                except Exception as e:
                    print(f"[MM] Failed to post ask: {e}")

        except Exception as e:
            print(f"[MM] Error updating quotes: {e}")

    def run(self, interval: int = 10):
        """Run the market maker loop"""
        print(f"[MM] Market Maker starting...")
        print(f"[MM] Spread: {self.spread}%, Size: {self.order_size}, Leverage: {self.leverage}x")
        print(f"[MM] Trader ID: {self.trader_id}")
        print(f"[MM] NOTE: All positions and leverage are PUBLIC")
        print()

        while True:
            self.update_quotes()
            time.sleep(interval)


def register_bot() -> tuple[str, str]:
    """Register a new market maker bot"""
    username = f"mm_bot_{random.randint(1000, 9999)}"
    resp = requests.post(
        f"{API_URL}/api/v1/auth/register",
        json={
            "username": username,
            "password": "bot_password",
            "type": "market_maker"
        }
    )
    resp.raise_for_status()
    data = resp.json()
    print(f"[MM] Registered as: {username}")
    return data["trader"]["id"], data["token"]


def main():
    print("=" * 50)
    print("  Trade.re Market Maker Bot")
    print("  Providing liquidity for R.index")
    print("=" * 50)
    print()

    # Register bot
    trader_id, token = register_bot()

    # Create and run market maker
    mm = MarketMaker(
        trader_id=trader_id,
        token=token,
        spread=0.3,      # 0.3% tighter spread for more fills
        order_size=3.0,  # 3 units per order
        leverage=5,      # Conservative 5x leverage
        num_levels=5     # 5 levels on each side for more depth
    )

    mm.run(interval=3)  # Update every 3 seconds for more activity


if __name__ == "__main__":
    main()
