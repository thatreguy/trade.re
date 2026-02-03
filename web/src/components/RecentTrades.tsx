'use client'

import { useEffect } from 'react'
import { useMarketStore } from '@/store/market'

export default function RecentTrades() {
  const { recentTrades, fetchRecentTrades } = useMarketStore()

  useEffect(() => {
    fetchRecentTrades()
    const interval = setInterval(fetchRecentTrades, 3000) // Refresh every 3s
    return () => clearInterval(interval)
  }, [fetchRecentTrades])

  const getLeverageBadge = (leverage: number) => {
    const tier = leverage <= 10 ? 'conservative' : leverage <= 50 ? 'moderate' : leverage <= 100 ? 'aggressive' : 'degen'
    return `leverage-${tier}`
  }

  const formatTime = (timestamp: string) => {
    try {
      const date = new Date(timestamp)
      return date.toLocaleTimeString('en-US', { hour12: false })
    } catch {
      return timestamp
    }
  }

  // Parse trade values from API strings to numbers
  const parsedTrades = recentTrades.map(trade => ({
    ...trade,
    price: parseFloat(trade.price) || 0,
    size: parseFloat(trade.size) || 0,
    buyer_leverage: trade.buyer_leverage || 0,
    seller_leverage: trade.seller_leverage || 0
  }))

  return (
    <div className="text-xs space-y-2 h-[220px] overflow-y-auto">
      {parsedTrades.length === 0 ? (
        <div className="text-center text-gray-500 py-4">
          No recent trades
        </div>
      ) : (
        parsedTrades.slice(0, 20).map((trade) => (
          <div key={trade.id} className="flex items-center justify-between py-1 border-b border-trade-border">
            <div>
              <span className={trade.aggressor_side === 'buy' ? 'text-trade-green' : 'text-trade-red'}>
                {trade.price.toFixed(2)}
              </span>
              <span className="text-gray-500 ml-2">{trade.size.toFixed(3)}</span>
            </div>
            <div className="flex items-center space-x-1">
              <span className={`px-1 rounded text-[10px] ${getLeverageBadge(trade.buyer_leverage)}`}>
                B:{trade.buyer_leverage}x
              </span>
              <span className={`px-1 rounded text-[10px] ${getLeverageBadge(trade.seller_leverage)}`}>
                S:{trade.seller_leverage}x
              </span>
              <span className="text-gray-600 ml-1">{formatTime(trade.timestamp)}</span>
            </div>
          </div>
        ))
      )}
    </div>
  )
}
