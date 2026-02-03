'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { useMarketStore } from '@/store/market'

export default function AllPositionsPage() {
  const { allPositions, fetchAllPositions } = useMarketStore()
  const [filter, setFilter] = useState<'all' | 'long' | 'short'>('all')
  const [leverageFilter, setLeverageFilter] = useState<string>('all')

  useEffect(() => {
    fetchAllPositions()
    const interval = setInterval(fetchAllPositions, 3000) // Refresh every 3s
    return () => clearInterval(interval)
  }, [fetchAllPositions])

  const positions = allPositions

  const getLeverageBadge = (leverage: number) => {
    const tier = leverage <= 10 ? 'conservative' : leverage <= 50 ? 'moderate' : leverage <= 100 ? 'aggressive' : 'degen'
    return `leverage-${tier}`
  }

  const filteredPositions = positions.filter(p => {
    if (filter === 'long' && p.size <= 0) return false
    if (filter === 'short' && p.size >= 0) return false
    if (leverageFilter === 'conservative' && p.leverage > 10) return false
    if (leverageFilter === 'moderate' && (p.leverage <= 10 || p.leverage > 50)) return false
    if (leverageFilter === 'aggressive' && (p.leverage <= 50 || p.leverage > 100)) return false
    if (leverageFilter === 'degen' && p.leverage <= 100) return false
    return true
  })

  const longPositions = positions.filter(p => p.size > 0)
  const shortPositions = positions.filter(p => p.size < 0)
  const totalLong = longPositions.reduce((sum, p) => sum + p.size, 0)
  const totalShort = Math.abs(shortPositions.reduce((sum, p) => sum + p.size, 0))
  const avgLongLeverage = longPositions.length > 0
    ? longPositions.reduce((sum, p) => sum + p.leverage, 0) / longPositions.length
    : 0
  const avgShortLeverage = shortPositions.length > 0
    ? shortPositions.reduce((sum, p) => sum + p.leverage, 0) / shortPositions.length
    : 0

  return (
    <div className="max-w-7xl mx-auto px-4 py-6">
      <h1 className="text-2xl font-bold mb-6">All Positions - Transparency Dashboard</h1>

      {/* Summary Stats */}
      <div className="grid grid-cols-4 gap-4 mb-6">
        <div className="bg-trade-card rounded-lg border border-trade-border p-4">
          <div className="text-gray-500 text-sm">Total Long</div>
          <div className="text-xl font-bold text-trade-green">{totalLong.toFixed(2)}</div>
          <div className="text-xs text-gray-500">Avg Leverage: {avgLongLeverage.toFixed(1)}x</div>
        </div>
        <div className="bg-trade-card rounded-lg border border-trade-border p-4">
          <div className="text-gray-500 text-sm">Total Short</div>
          <div className="text-xl font-bold text-trade-red">{totalShort.toFixed(2)}</div>
          <div className="text-xs text-gray-500">Avg Leverage: {avgShortLeverage.toFixed(1)}x</div>
        </div>
        <div className="bg-trade-card rounded-lg border border-trade-border p-4">
          <div className="text-gray-500 text-sm">Long/Short Ratio</div>
          <div className="text-xl font-bold">{(totalLong / totalShort).toFixed(2)}</div>
        </div>
        <div className="bg-trade-card rounded-lg border border-trade-border p-4">
          <div className="text-gray-500 text-sm">Open Positions</div>
          <div className="text-xl font-bold">{positions.length}</div>
        </div>
      </div>

      {/* Filters */}
      <div className="flex space-x-4 mb-4">
        <div className="flex space-x-2">
          <button
            className={`px-4 py-2 rounded ${filter === 'all' ? 'bg-white text-black' : 'bg-trade-card text-gray-400'}`}
            onClick={() => setFilter('all')}
          >
            All
          </button>
          <button
            className={`px-4 py-2 rounded ${filter === 'long' ? 'bg-trade-green text-white' : 'bg-trade-card text-gray-400'}`}
            onClick={() => setFilter('long')}
          >
            Longs
          </button>
          <button
            className={`px-4 py-2 rounded ${filter === 'short' ? 'bg-trade-red text-white' : 'bg-trade-card text-gray-400'}`}
            onClick={() => setFilter('short')}
          >
            Shorts
          </button>
        </div>

        <select
          value={leverageFilter}
          onChange={(e) => setLeverageFilter(e.target.value)}
          className="bg-trade-card border border-trade-border rounded px-4 py-2 text-white"
        >
          <option value="all">All Leverage</option>
          <option value="conservative">Conservative (1-10x)</option>
          <option value="moderate">Moderate (11-50x)</option>
          <option value="aggressive">Aggressive (51-100x)</option>
          <option value="degen">Degen (101-150x)</option>
        </select>
      </div>

      {/* Positions Table */}
      <div className="bg-trade-card rounded-lg border border-trade-border overflow-hidden">
        <table className="w-full">
          <thead className="bg-trade-bg">
            <tr className="text-left text-sm text-gray-500">
              <th className="px-4 py-3">Trader</th>
              <th className="px-4 py-3">Side</th>
              <th className="px-4 py-3">Size</th>
              <th className="px-4 py-3">Entry</th>
              <th className="px-4 py-3">Leverage</th>
              <th className="px-4 py-3">Margin</th>
              <th className="px-4 py-3">Unrealized P&L</th>
              <th className="px-4 py-3">Liq. Price</th>
            </tr>
          </thead>
          <tbody>
            {filteredPositions.length === 0 ? (
              <tr>
                <td colSpan={8} className="px-4 py-8 text-center text-gray-500">
                  No open positions
                </td>
              </tr>
            ) : (
              filteredPositions.map((pos) => (
                <tr key={pos.trader_id} className="border-t border-trade-border hover:bg-trade-bg/50">
                  <td className="px-4 py-3 font-medium">
                    <Link href={`/trader/${pos.trader_id}`} className="hover:text-blue-400">
                      {pos.trader_id.slice(0, 8)}...
                    </Link>
                  </td>
                  <td className="px-4 py-3">
                    <span className={pos.size > 0 ? 'text-trade-green' : 'text-trade-red'}>
                      {pos.size > 0 ? 'LONG' : 'SHORT'}
                    </span>
                  </td>
                  <td className="px-4 py-3">{Math.abs(pos.size).toFixed(3)}</td>
                  <td className="px-4 py-3">${pos.entry_price.toFixed(2)}</td>
                  <td className="px-4 py-3">
                    <span className={`px-2 py-1 rounded text-xs ${getLeverageBadge(pos.leverage)}`}>
                      {pos.leverage}x
                    </span>
                  </td>
                  <td className="px-4 py-3">${pos.margin.toFixed(2)}</td>
                  <td className={`px-4 py-3 ${pos.unrealized_pnl >= 0 ? 'text-trade-green' : 'text-trade-red'}`}>
                    {pos.unrealized_pnl >= 0 ? '+' : ''}${pos.unrealized_pnl.toFixed(2)}
                  </td>
                  <td className="px-4 py-3 text-trade-red">${pos.liquidation_price.toFixed(2)}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Transparency Notice */}
      <div className="mt-6 text-center text-gray-500 text-sm">
        All position data is updated in real-time. Everyone sees the same information.
      </div>
    </div>
  )
}
