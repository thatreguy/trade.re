'use client'

import { useEffect, useState } from 'react'

interface Position {
  traderId: string
  username: string
  traderType: string
  size: number
  entryPrice: number
  leverage: number
  margin: number
  unrealizedPnl: number
  liquidationPrice: number
}

export default function AllPositionsPage() {
  const [positions, setPositions] = useState<Position[]>([])
  const [filter, setFilter] = useState<'all' | 'long' | 'short'>('all')
  const [leverageFilter, setLeverageFilter] = useState<string>('all')

  // Mock data
  useEffect(() => {
    setPositions([
      { traderId: '1', username: 'whale_trader', traderType: 'human', size: 100, entryPrice: 995.00, leverage: 50, margin: 1990, unrealizedPnl: 575, liquidationPrice: 975.10 },
      { traderId: '2', username: 'mm_bot_1', traderType: 'market_maker', size: -75, entryPrice: 1002.00, leverage: 10, margin: 7515, unrealizedPnl: -112.50, liquidationPrice: 1101.20 },
      { traderId: '3', username: 'degen_andy', traderType: 'human', size: 25, entryPrice: 998.00, leverage: 150, margin: 166.33, unrealizedPnl: 68.75, liquidationPrice: 991.35 },
      { traderId: '4', username: 'algo_sniper', traderType: 'bot', size: -50, entryPrice: 1001.50, leverage: 25, margin: 2003, unrealizedPnl: 37.50, liquidationPrice: 1041.16 },
      { traderId: '5', username: 'conservative_carl', traderType: 'human', size: 10, entryPrice: 990.00, leverage: 3, margin: 3300, unrealizedPnl: 107.50, liquidationPrice: 660.00 },
      { traderId: '6', username: 'trend_follower', traderType: 'bot', size: 30, entryPrice: 1000.00, leverage: 75, margin: 400, unrealizedPnl: 22.50, liquidationPrice: 986.67 },
    ])
  }, [])

  const getLeverageBadge = (leverage: number) => {
    const tier = leverage <= 10 ? 'conservative' : leverage <= 50 ? 'moderate' : leverage <= 100 ? 'aggressive' : 'degen'
    return `leverage-${tier}`
  }

  const getTraderTypeBadge = (type: string) => {
    switch (type) {
      case 'human': return 'bg-blue-600'
      case 'bot': return 'bg-orange-600'
      case 'market_maker': return 'bg-purple-600'
      default: return 'bg-gray-600'
    }
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

  const totalLong = positions.filter(p => p.size > 0).reduce((sum, p) => sum + p.size, 0)
  const totalShort = Math.abs(positions.filter(p => p.size < 0).reduce((sum, p) => sum + p.size, 0))
  const avgLongLeverage = positions.filter(p => p.size > 0).reduce((sum, p) => sum + p.leverage, 0) / positions.filter(p => p.size > 0).length || 0
  const avgShortLeverage = positions.filter(p => p.size < 0).reduce((sum, p) => sum + p.leverage, 0) / positions.filter(p => p.size < 0).length || 0

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
              <th className="px-4 py-3">Type</th>
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
            {filteredPositions.map((pos) => (
              <tr key={pos.traderId} className="border-t border-trade-border hover:bg-trade-bg/50">
                <td className="px-4 py-3 font-medium">{pos.username}</td>
                <td className="px-4 py-3">
                  <span className={`px-2 py-1 rounded text-xs ${getTraderTypeBadge(pos.traderType)}`}>
                    {pos.traderType}
                  </span>
                </td>
                <td className="px-4 py-3">
                  <span className={pos.size > 0 ? 'text-trade-green' : 'text-trade-red'}>
                    {pos.size > 0 ? 'LONG' : 'SHORT'}
                  </span>
                </td>
                <td className="px-4 py-3">{Math.abs(pos.size).toFixed(2)}</td>
                <td className="px-4 py-3">${pos.entryPrice.toFixed(2)}</td>
                <td className="px-4 py-3">
                  <span className={`px-2 py-1 rounded text-xs ${getLeverageBadge(pos.leverage)}`}>
                    {pos.leverage}x
                  </span>
                </td>
                <td className="px-4 py-3">${pos.margin.toFixed(2)}</td>
                <td className={`px-4 py-3 ${pos.unrealizedPnl >= 0 ? 'text-trade-green' : 'text-trade-red'}`}>
                  {pos.unrealizedPnl >= 0 ? '+' : ''}${pos.unrealizedPnl.toFixed(2)}
                </td>
                <td className="px-4 py-3 text-trade-red">${pos.liquidationPrice.toFixed(2)}</td>
              </tr>
            ))}
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
