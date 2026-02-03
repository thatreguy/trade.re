'use client'

import { useState } from 'react'
import { useUserStore } from '@/store/user'
import { useMarketStore } from '@/store/market'
import { api } from '@/lib/api'

export default function PositionPanel() {
  const { position, isAuthenticated, fetchUserData } = useUserStore()
  const { marketStats } = useMarketStore()
  const [isClosing, setIsClosing] = useState(false)

  // Not logged in
  if (!isAuthenticated) {
    return (
      <div className="text-gray-500 text-sm text-center py-4">
        Login to see your position
      </div>
    )
  }

  // No position
  if (!position || position.size === 0) {
    return (
      <div className="text-gray-500 text-sm text-center py-4">
        No open position
      </div>
    )
  }

  const isLong = position.size > 0
  const markPrice = marketStats?.mark_price || marketStats?.last_price || position.entry_price
  const pnlPercent = position.margin > 0 ? ((position.unrealized_pnl / position.margin) * 100).toFixed(2) : '0.00'
  const leverageTier = position.leverage <= 10 ? 'conservative' : position.leverage <= 50 ? 'moderate' : position.leverage <= 100 ? 'aggressive' : 'degen'

  const handleClosePosition = async () => {
    setIsClosing(true)
    try {
      await api.closePosition()
      await fetchUserData()
    } catch (e) {
      console.error('Failed to close position:', e)
    } finally {
      setIsClosing(false)
    }
  }

  return (
    <div className="text-sm">
      {/* Position Header */}
      <div className="flex items-center justify-between mb-3">
        <span className={`font-bold ${isLong ? 'text-trade-green' : 'text-trade-red'}`}>
          {isLong ? 'LONG' : 'SHORT'} {Math.abs(position.size).toFixed(3)}
        </span>
        <span className={`px-2 py-0.5 rounded text-xs leverage-${leverageTier}`}>
          {position.leverage}x
        </span>
      </div>

      {/* Position Details */}
      <div className="space-y-2">
        <div className="flex justify-between">
          <span className="text-gray-500">Entry Price</span>
          <span>${position.entry_price.toFixed(2)}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-gray-500">Mark Price</span>
          <span>${markPrice.toFixed(2)}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-gray-500">Margin</span>
          <span>${position.margin.toFixed(2)}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-gray-500">Unrealized P&L</span>
          <span className={position.unrealized_pnl >= 0 ? 'text-trade-green' : 'text-trade-red'}>
            {position.unrealized_pnl >= 0 ? '+' : ''}${position.unrealized_pnl.toFixed(2)} ({pnlPercent}%)
          </span>
        </div>
        <div className="flex justify-between">
          <span className="text-gray-500">Liq. Price</span>
          <span className="text-trade-red">${position.liquidation_price.toFixed(2)}</span>
        </div>
      </div>

      {/* Close Button */}
      <button
        onClick={handleClosePosition}
        disabled={isClosing}
        className="w-full mt-4 py-2 bg-trade-border hover:bg-gray-600 rounded text-sm disabled:opacity-50"
      >
        {isClosing ? 'Closing...' : 'Close Position'}
      </button>
    </div>
  )
}
