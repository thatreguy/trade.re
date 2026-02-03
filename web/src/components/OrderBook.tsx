'use client'

import { useEffect, useState } from 'react'

interface OrderBookLevel {
  price: string
  size: string
  orderCount: number
}

interface OrderBookData {
  bids: OrderBookLevel[]
  asks: OrderBookLevel[]
}

export default function OrderBook() {
  const [orderBook, setOrderBook] = useState<OrderBookData>({ bids: [], asks: [] })

  // Mock data for demonstration
  useEffect(() => {
    setOrderBook({
      asks: [
        { price: '1005.50', size: '12.5', orderCount: 3 },
        { price: '1004.00', size: '8.2', orderCount: 2 },
        { price: '1003.25', size: '15.0', orderCount: 4 },
        { price: '1002.00', size: '5.8', orderCount: 1 },
        { price: '1001.50', size: '22.3', orderCount: 5 },
      ],
      bids: [
        { price: '1000.00', size: '18.7', orderCount: 4 },
        { price: '999.50', size: '10.2', orderCount: 2 },
        { price: '998.75', size: '25.0', orderCount: 6 },
        { price: '998.00', size: '8.5', orderCount: 2 },
        { price: '997.25', size: '14.3', orderCount: 3 },
      ],
    })
  }, [])

  const maxSize = Math.max(
    ...orderBook.bids.map(b => parseFloat(b.size)),
    ...orderBook.asks.map(a => parseFloat(a.size)),
    1
  )

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
        {orderBook.asks.slice().reverse().map((ask, i) => (
          <div key={i} className="grid grid-cols-3 relative">
            <div
              className="absolute inset-0 bg-red-500/20"
              style={{ width: `${(parseFloat(ask.size) / maxSize) * 100}%`, right: 0, left: 'auto' }}
            />
            <span className="relative text-trade-red">{ask.price}</span>
            <span className="relative text-right">{ask.size}</span>
            <span className="relative text-right text-gray-500">{ask.orderCount}</span>
          </div>
        ))}
      </div>

      {/* Spread */}
      <div className="my-2 py-2 border-y border-trade-border text-center">
        <span className="text-lg font-bold">1000.75</span>
        <span className="text-xs text-gray-500 ml-2">Spread: 0.50</span>
      </div>

      {/* Bids */}
      <div className="space-y-1">
        {orderBook.bids.map((bid, i) => (
          <div key={i} className="grid grid-cols-3 relative">
            <div
              className="absolute inset-0 bg-green-500/20"
              style={{ width: `${(parseFloat(bid.size) / maxSize) * 100}%` }}
            />
            <span className="relative text-trade-green">{bid.price}</span>
            <span className="relative text-right">{bid.size}</span>
            <span className="relative text-right text-gray-500">{bid.orderCount}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
