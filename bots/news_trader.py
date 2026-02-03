#!/usr/bin/env python3
"""
News Sentiment Trading Bot for Trade.re

This bot curates world news headlines and takes bullish/bearish positions
based on sentiment analysis. All positions are PUBLIC - transparency is key!

Uses simple keyword-based sentiment as a demo. In production, you'd use
a proper NLP model or sentiment API.

Usage:
    python news_trader.py

Environment variables:
    API_URL: Backend API URL (default: http://localhost:8080)
    NEWS_API_KEY: NewsAPI.org API key (optional, uses mock data if not set)
"""

import os
import time
import json
import random
import requests
from typing import Optional, List
from dataclasses import dataclass
from datetime import datetime

API_URL = os.getenv("API_URL", "http://localhost:8080")
NEWS_API_KEY = os.getenv("NEWS_API_KEY", "")

# Sentiment keywords (simplified)
BULLISH_KEYWORDS = [
    "surge", "soar", "rally", "gain", "growth", "bullish", "record high",
    "boom", "recovery", "optimism", "breakthrough", "success", "profit",
    "strong", "up", "rise", "positive", "confident", "buy"
]

BEARISH_KEYWORDS = [
    "crash", "plunge", "fall", "drop", "decline", "bearish", "record low",
    "recession", "fear", "concern", "crisis", "loss", "weak", "down",
    "negative", "sell", "panic", "warning", "risk", "uncertainty"
]

# Mock news for when no API key is provided
MOCK_NEWS = [
    {"title": "Markets surge on positive economic data", "sentiment": "bullish"},
    {"title": "Tech stocks rally as earnings beat expectations", "sentiment": "bullish"},
    {"title": "Concerns grow over inflation data", "sentiment": "bearish"},
    {"title": "Strong jobs report boosts market confidence", "sentiment": "bullish"},
    {"title": "Oil prices drop amid demand worries", "sentiment": "bearish"},
    {"title": "Central bank signals optimism for growth", "sentiment": "bullish"},
    {"title": "Trade tensions rise between major economies", "sentiment": "bearish"},
    {"title": "Consumer spending shows strong growth", "sentiment": "bullish"},
    {"title": "Market volatility increases on uncertainty", "sentiment": "bearish"},
    {"title": "Innovation breakthroughs drive tech sector", "sentiment": "bullish"},
]


@dataclass
class NewsTrader:
    trader_id: str
    token: str
    position_size: float = 2.0  # Size per trade
    leverage: int = 25  # Moderate leverage
    sentiment_threshold: float = 0.3  # Min sentiment score to act

    def get_market_stats(self) -> dict:
        """Get current market stats"""
        resp = requests.get(f"{API_URL}/api/v1/market/stats")
        resp.raise_for_status()
        return resp.json()

    def get_position(self) -> Optional[dict]:
        """Get current position"""
        resp = requests.get(f"{API_URL}/api/v1/traders/{self.trader_id}/positions")
        resp.raise_for_status()
        positions = resp.json()
        return positions[0] if positions else None

    def place_order(self, side: str, size: float) -> dict:
        """Place a market order"""
        resp = requests.post(
            f"{API_URL}/api/v1/orders",
            json={
                "trader_id": self.trader_id,
                "instrument": "R.index",
                "side": side,
                "type": "market",
                "price": "0",  # Market order
                "size": str(size),
                "leverage": self.leverage
            },
            headers={"Authorization": f"Bearer {self.token}"}
        )
        resp.raise_for_status()
        return resp.json()

    def fetch_news(self) -> List[dict]:
        """Fetch news headlines"""
        if NEWS_API_KEY:
            try:
                resp = requests.get(
                    "https://newsapi.org/v2/top-headlines",
                    params={
                        "apiKey": NEWS_API_KEY,
                        "category": "business",
                        "language": "en",
                        "pageSize": 10
                    }
                )
                resp.raise_for_status()
                return resp.json().get("articles", [])
            except Exception as e:
                print(f"[NEWS] Error fetching news: {e}")
                return []
        else:
            # Use mock news
            return random.sample(MOCK_NEWS, min(5, len(MOCK_NEWS)))

    def analyze_sentiment(self, headlines: List[str]) -> float:
        """
        Analyze sentiment from headlines.
        Returns a score from -1 (very bearish) to +1 (very bullish)
        """
        if not headlines:
            return 0.0

        bullish_count = 0
        bearish_count = 0

        for headline in headlines:
            headline_lower = headline.lower()
            for keyword in BULLISH_KEYWORDS:
                if keyword in headline_lower:
                    bullish_count += 1
            for keyword in BEARISH_KEYWORDS:
                if keyword in headline_lower:
                    bearish_count += 1

        total = bullish_count + bearish_count
        if total == 0:
            return 0.0

        return (bullish_count - bearish_count) / total

    def trade_on_sentiment(self):
        """Analyze news and trade based on sentiment"""
        try:
            # Fetch and analyze news
            news = self.fetch_news()
            headlines = [
                article.get("title", article.get("headline", ""))
                for article in news
            ]

            if not headlines:
                print("[NEWS] No headlines to analyze")
                return

            sentiment = self.analyze_sentiment(headlines)
            print(f"[NEWS] Headlines analyzed: {len(headlines)}")
            print(f"[NEWS] Sentiment score: {sentiment:.2f}")

            # Log headlines
            for h in headlines[:3]:
                print(f"[NEWS]   - {h[:60]}...")

            # Only trade if sentiment is strong enough
            if abs(sentiment) < self.sentiment_threshold:
                print(f"[NEWS] Sentiment too weak ({sentiment:.2f}), no trade")
                return

            # Get current position
            position = self.get_position()
            current_size = float(position["size"]) if position else 0

            if sentiment > self.sentiment_threshold:
                # Bullish - go long or add to long
                if current_size >= 0:
                    print(f"[NEWS] BULLISH signal! Going LONG {self.position_size} @ {self.leverage}x")
                    self.place_order("buy", self.position_size)
                else:
                    print(f"[NEWS] BULLISH signal but already SHORT, holding")

            elif sentiment < -self.sentiment_threshold:
                # Bearish - go short or add to short
                if current_size <= 0:
                    print(f"[NEWS] BEARISH signal! Going SHORT {self.position_size} @ {self.leverage}x")
                    self.place_order("sell", self.position_size)
                else:
                    print(f"[NEWS] BEARISH signal but already LONG, holding")

        except Exception as e:
            print(f"[NEWS] Error in trade_on_sentiment: {e}")

    def run(self, interval: int = 60):
        """Run the news trader loop"""
        print(f"[NEWS] News Sentiment Trader starting...")
        print(f"[NEWS] Position size: {self.position_size}, Leverage: {self.leverage}x")
        print(f"[NEWS] Sentiment threshold: {self.sentiment_threshold}")
        print(f"[NEWS] Trader ID: {self.trader_id}")
        print(f"[NEWS] NOTE: All positions and leverage are PUBLIC")
        if not NEWS_API_KEY:
            print(f"[NEWS] Using mock news (set NEWS_API_KEY for real news)")
        print()

        while True:
            print(f"\n[NEWS] {datetime.now().strftime('%H:%M:%S')} - Checking news...")
            self.trade_on_sentiment()
            time.sleep(interval)


def register_bot() -> tuple[str, str]:
    """Register a new trading bot"""
    username = f"news_bot_{random.randint(1000, 9999)}"
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
    print(f"[NEWS] Registered as: {username}")
    return data["trader"]["id"], data["token"]


def main():
    print("=" * 50)
    print("  Trade.re News Sentiment Trading Bot")
    print("  Curating world news for R.index trading")
    print("=" * 50)
    print()

    # Register bot
    trader_id, token = register_bot()

    # Create and run news trader
    trader = NewsTrader(
        trader_id=trader_id,
        token=token,
        position_size=2.0,       # 2 units per trade
        leverage=25,             # Moderate 25x leverage
        sentiment_threshold=0.3  # Need 30% sentiment bias to trade
    )

    trader.run(interval=30)  # Check news every 30 seconds


if __name__ == "__main__":
    main()
