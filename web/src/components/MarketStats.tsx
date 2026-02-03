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

  // Parse string values from API to numbers
  const lastPrice = parseFloat(marketStats?.last_price) || 1000
  const high24h = parseFloat(marketStats?.high_24h) || 1000
  const low24h = parseFloat(marketStats?.low_24h) || 1000
  const volume24h = parseFloat(marketStats?.volume_24h) || 0
  const openInterest = parseFloat(marketStats?.open_interest) || 0
  const insuranceFund = parseFloat(marketStats?.insurance_fund) || 1000000

  const change24h = low24h > 0 ? ((lastPrice - low24h) / low24h * 100) : 0
  const isPositive = change24h >= 0

  return (
    <div className="bg-trade-card rounded-lg border border-trade-border p-4">
      <div className="flex items-center justify-between">
        {/* Price & Change */}
        <div className="flex items-center space-x-6">
          <div>
            <div className="text-2xl font-bold">
              ${lastPrice.toFixed(2)}
            </div>
            <div className={`text-sm ${isPositive ? 'text-trade-green' : 'text-trade-red'}`}>
              {isPositive ? '+' : ''}{change24h.toFixed(2)}%
            </div>
          </div>

          <div className="h-10 border-l border-trade-border" />

          <div className="text-sm">
            <div className="text-gray-500">24h High</div>
            <div>${high24h.toFixed(2)}</div>
          </div>

          <div className="text-sm">
            <div className="text-gray-500">24h Low</div>
            <div>${low24h.toFixed(2)}</div>
          </div>

          <div className="text-sm">
            <div className="text-gray-500">24h Volume</div>
            <div>${(volume24h / 1000).toFixed(1)}K</div>
          </div>
        </div>

        {/* OI & Insurance */}
        <div className="flex items-center space-x-6">
          <div className="text-sm">
            <div className="text-gray-500">Open Interest</div>
            <div>${(openInterest / 1000).toFixed(1)}K</div>
          </div>

          <div className="text-sm">
            <div className="text-gray-500">Insurance Fund</div>
            <div className="text-purple-400">${(insuranceFund / 1000000).toFixed(2)}M</div>
          </div>

          <div className="text-xs text-gray-500">
            24/7 Market
          </div>
        </div>
      </div>
    </div>
  )
}
