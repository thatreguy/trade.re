'use client'

import { useEffect } from 'react'
import { useMarketStore } from '@/store/market'

export default function MarketStats() {
  const { marketStats, fetchMarketStats } = useMarketStore()

  useEffect(() => {
    fetchMarketStats()
    const interval = setInterval(fetchMarketStats, 5000) // Refresh every 5s
    return () => clearInterval(interval)
  }, [fetchMarketStats])

  // Default values while loading
  const stats = marketStats || {
    last_price: 1000,
    high_24h: 1000,
    low_24h: 1000,
    volume_24h: 0,
    open_interest: 0,
    insurance_fund: 1000000,
  }

  const change24h = stats.high_24h > 0
    ? ((stats.last_price - stats.low_24h) / stats.low_24h * 100)
    : 0
  const isPositive = change24h >= 0

  return (
    <div className="bg-trade-card rounded-lg border border-trade-border p-4">
      <div className="flex items-center justify-between">
        {/* Price & Change */}
        <div className="flex items-center space-x-6">
          <div>
            <div className="text-2xl font-bold">
              ${stats.last_price.toFixed(2)}
            </div>
            <div className={`text-sm ${isPositive ? 'text-trade-green' : 'text-trade-red'}`}>
              {isPositive ? '+' : ''}{change24h.toFixed(2)}%
            </div>
          </div>

          <div className="h-10 border-l border-trade-border" />

          <div className="text-sm">
            <div className="text-gray-500">24h High</div>
            <div>${stats.high_24h.toFixed(2)}</div>
          </div>

          <div className="text-sm">
            <div className="text-gray-500">24h Low</div>
            <div>${stats.low_24h.toFixed(2)}</div>
          </div>

          <div className="text-sm">
            <div className="text-gray-500">24h Volume</div>
            <div>${(stats.volume_24h / 1000).toFixed(1)}K</div>
          </div>
        </div>

        {/* OI & Insurance */}
        <div className="flex items-center space-x-6">
          <div className="text-sm">
            <div className="text-gray-500">Open Interest</div>
            <div>${(stats.open_interest / 1000).toFixed(1)}K</div>
          </div>

          <div className="text-sm">
            <div className="text-gray-500">Insurance Fund</div>
            <div className="text-purple-400">${(stats.insurance_fund / 1000000).toFixed(2)}M</div>
          </div>

          <div className="text-xs text-gray-500">
            24/7 Market
          </div>
        </div>
      </div>
    </div>
  )
}
