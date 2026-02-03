'use client'

import { useEffect, useState } from 'react'

interface Trade {
  id: string
  price: string
  size: string
  side: 'buy' | 'sell'
  buyerLeverage: number
  sellerLeverage: number
  timestamp: string
}

export default function RecentTrades() {
  const [trades, setTrades] = useState<Trade[]>([])

  // Mock data
  useEffect(() => {
    setTrades([
      { id: '1', price: '1000.50', size: '2.5', side: 'buy', buyerLeverage: 50, sellerLeverage: 10, timestamp: '12:34:56' },
      { id: '2', price: '1000.25', size: '1.0', side: 'sell', buyerLeverage: 25, sellerLeverage: 100, timestamp: '12:34:52' },
      { id: '3', price: '1000.75', size: '5.0', side: 'buy', buyerLeverage: 10, sellerLeverage: 5, timestamp: '12:34:48' },
      { id: '4', price: '999.50', size: '3.2', side: 'sell', buyerLeverage: 75, sellerLeverage: 20, timestamp: '12:34:45' },
      { id: '5', price: '1001.00', size: '0.8', side: 'buy', buyerLeverage: 150, sellerLeverage: 3, timestamp: '12:34:40' },
    ])
  }, [])

  const getLeverageBadge = (leverage: number) => {
    const tier = leverage <= 10 ? 'conservative' : leverage <= 50 ? 'moderate' : leverage <= 100 ? 'aggressive' : 'degen'
    return `leverage-${tier}`
  }

  return (
    <div className="text-xs space-y-2 max-h-64 overflow-y-auto">
      {trades.map((trade) => (
        <div key={trade.id} className="flex items-center justify-between py-1 border-b border-trade-border">
          <div>
            <span className={trade.side === 'buy' ? 'text-trade-green' : 'text-trade-red'}>
              {trade.price}
            </span>
            <span className="text-gray-500 ml-2">{trade.size}</span>
          </div>
          <div className="flex items-center space-x-1">
            <span className={`px-1 rounded text-[10px] ${getLeverageBadge(trade.buyerLeverage)}`}>
              B:{trade.buyerLeverage}x
            </span>
            <span className={`px-1 rounded text-[10px] ${getLeverageBadge(trade.sellerLeverage)}`}>
              S:{trade.sellerLeverage}x
            </span>
          </div>
        </div>
      ))}
    </div>
  )
}
