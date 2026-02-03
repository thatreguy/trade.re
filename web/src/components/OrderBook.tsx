'use client'

import { useEffect } from 'react'
import { useMarketStore } from '@/store/market'

export default function OrderBook() {
  const { orderBook, fetchOrderBook } = useMarketStore()

  useEffect(() => {
    fetchOrderBook()
    const interval = setInterval(fetchOrderBook, 2000) // Refresh every 2s
    return () => clearInterval(interval)
  }, [fetchOrderBook])

  // Parse string values from API to numbers
  const bids = (orderBook?.bids || []).map(b => ({
    price: parseFloat(b.price) || 0,
    size: parseFloat(b.size) || 0,
    order_count: b.order_count || 0
  }))
  const asks = (orderBook?.asks || []).map(a => ({
    price: parseFloat(a.price) || 0,
    size: parseFloat(a.size) || 0,
    order_count: a.order_count || 0
  }))

  const maxSize = Math.max(
    ...bids.map(b => b.size),
    ...asks.map(a => a.size),
    1
  )

  // Calculate spread and mid price
  const bestBid = bids[0]?.price || 0
  const bestAsk = asks[0]?.price || 0
  const spread = bestAsk > 0 && bestBid > 0 ? (bestAsk - bestBid).toFixed(2) : '—'
  const midPrice = bestAsk > 0 && bestBid > 0 ? ((bestAsk + bestBid) / 2).toFixed(2) : '—'

  return (
    <div className="text-sm">
      {/* Header */}
      <div className="grid grid-cols-3 text-xs text-gray-500 mb-2">
        <span>Price</span>
        <span className="text-right">Size</span>
        <span className="text-right">Orders</span>
      </div>

      {/* Asks (reversed so lowest ask is at bottom) */}
      <div className="space-y-1">
        {asks.slice(0, 5).reverse().map((ask, i) => (
          <div key={i} className="grid grid-cols-3 relative">
            <div
              className="absolute inset-0 bg-red-500/20"
              style={{ width: `${(ask.size / maxSize) * 100}%`, right: 0, left: 'auto' }}
            />
            <span className="relative text-trade-red">{ask.price.toFixed(2)}</span>
            <span className="relative text-right">{ask.size.toFixed(2)}</span>
            <span className="relative text-right text-gray-500">{ask.order_count}</span>
          </div>
        ))}
      </div>

      {/* Spread */}
      <div className="my-2 py-2 border-y border-trade-border text-center">
        <span className="text-lg font-bold">{midPrice}</span>
        <span className="text-xs text-gray-500 ml-2">Spread: {spread}</span>
      </div>

      {/* Bids */}
      <div className="space-y-1">
        {bids.slice(0, 5).map((bid, i) => (
          <div key={i} className="grid grid-cols-3 relative">
            <div
              className="absolute inset-0 bg-green-500/20"
              style={{ width: `${(bid.size / maxSize) * 100}%` }}
            />
            <span className="relative text-trade-green">{bid.price.toFixed(2)}</span>
            <span className="relative text-right">{bid.size.toFixed(2)}</span>
            <span className="relative text-right text-gray-500">{bid.order_count}</span>
          </div>
        ))}
      </div>

      {bids.length === 0 && asks.length === 0 && (
        <div className="text-center text-gray-500 py-4">
          No orders in book
        </div>
      )}
    </div>
  )
}
