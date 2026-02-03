'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { useMarketStore } from '@/store/market'
import { Trader } from '@/lib/api'

export default function LeaderboardPage() {
  const { traders, fetchTraders } = useMarketStore()
  const [sortBy, setSortBy] = useState<'pnl' | 'trades' | 'balance'>('pnl')
  const [leverageFilter, setLeverageFilter] = useState<string>('all')
  const [typeFilter, setTypeFilter] = useState<string>('all')

  useEffect(() => {
    fetchTraders()
    const interval = setInterval(fetchTraders, 10000) // Refresh every 10s
    return () => clearInterval(interval)
  }, [fetchTraders])

  // Parse trader values from API strings to numbers
  const parsedTraders = traders.map(t => ({
    ...t,
    balance: parseFloat(t.balance) || 0,
    total_pnl: parseFloat(t.total_pnl) || 0,
    trade_count: t.trade_count || 0,
    max_leverage_used: t.max_leverage_used || 0
  }))

  const getLeverageTier = (leverage: number) => {
    if (leverage <= 10) return 'conservative'
    if (leverage <= 50) return 'moderate'
    if (leverage <= 100) return 'aggressive'
    return 'degen'
  }

  const getLeverageBadge = (leverage: number) => {
    return `leverage-${getLeverageTier(leverage)}`
  }

  const getTraderTypeBadge = (type: string) => {
    switch (type) {
      case 'human': return 'bg-blue-600'
      case 'bot': return 'bg-orange-600'
      case 'market_maker': return 'bg-purple-600'
      default: return 'bg-gray-600'
    }
  }

  const filteredTraders = parsedTraders
    .filter(t => {
      if (typeFilter !== 'all' && t.type !== typeFilter) return false
      if (leverageFilter !== 'all' && getLeverageTier(t.max_leverage_used) !== leverageFilter) return false
      return true
    })
    .sort((a, b) => {
      switch (sortBy) {
        case 'pnl': return b.total_pnl - a.total_pnl
        case 'trades': return b.trade_count - a.trade_count
        case 'balance': return b.balance - a.balance
        default: return 0
      }
    })

  const totalPnlPositive = parsedTraders.filter(t => t.total_pnl > 0).length
  const totalPnlNegative = parsedTraders.filter(t => t.total_pnl < 0).length
  const avgLeverage = parsedTraders.length > 0
    ? parsedTraders.reduce((sum, t) => sum + t.max_leverage_used, 0) / parsedTraders.length
    : 0

  return (
    <div className="max-w-7xl mx-auto px-4 py-6">
      <h1 className="text-2xl font-bold mb-2">Leaderboard</h1>
      <p className="text-gray-500 mb-6">Rankings are public. See who's winning and at what leverage.</p>

      {/* Summary Stats */}
      <div className="grid grid-cols-4 gap-4 mb-6">
        <div className="bg-trade-card rounded-lg border border-trade-border p-4">
          <div className="text-gray-500 text-sm">Total Traders</div>
          <div className="text-xl font-bold">{traders.length}</div>
        </div>
        <div className="bg-trade-card rounded-lg border border-trade-border p-4">
          <div className="text-gray-500 text-sm">Profitable</div>
          <div className="text-xl font-bold text-trade-green">{totalPnlPositive}</div>
        </div>
        <div className="bg-trade-card rounded-lg border border-trade-border p-4">
          <div className="text-gray-500 text-sm">Losing</div>
          <div className="text-xl font-bold text-trade-red">{totalPnlNegative}</div>
        </div>
        <div className="bg-trade-card rounded-lg border border-trade-border p-4">
          <div className="text-gray-500 text-sm">Avg Max Leverage</div>
          <div className="text-xl font-bold">{avgLeverage.toFixed(0)}x</div>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap gap-4 mb-6">
        <div className="flex items-center space-x-2">
          <span className="text-gray-500 text-sm">Sort by:</span>
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as any)}
            className="bg-trade-card border border-trade-border rounded px-3 py-2 text-sm"
          >
            <option value="pnl">Total P&L</option>
            <option value="trades">Trade Count</option>
            <option value="balance">Balance</option>
          </select>
        </div>

        <div className="flex items-center space-x-2">
          <span className="text-gray-500 text-sm">Leverage:</span>
          <select
            value={leverageFilter}
            onChange={(e) => setLeverageFilter(e.target.value)}
            className="bg-trade-card border border-trade-border rounded px-3 py-2 text-sm"
          >
            <option value="all">All</option>
            <option value="conservative">Conservative (1-10x)</option>
            <option value="moderate">Moderate (11-50x)</option>
            <option value="aggressive">Aggressive (51-100x)</option>
            <option value="degen">Degen (101-150x)</option>
          </select>
        </div>

        <div className="flex items-center space-x-2">
          <span className="text-gray-500 text-sm">Type:</span>
          <select
            value={typeFilter}
            onChange={(e) => setTypeFilter(e.target.value)}
            className="bg-trade-card border border-trade-border rounded px-3 py-2 text-sm"
          >
            <option value="all">All</option>
            <option value="human">Human</option>
            <option value="bot">Bot</option>
            <option value="market_maker">Market Maker</option>
          </select>
        </div>
      </div>

      {/* Leaderboard Table */}
      <div className="bg-trade-card rounded-lg border border-trade-border overflow-hidden">
        <table className="w-full">
          <thead className="bg-trade-bg">
            <tr className="text-left text-sm text-gray-500">
              <th className="px-4 py-3 w-12">#</th>
              <th className="px-4 py-3">Trader</th>
              <th className="px-4 py-3">Type</th>
              <th className="px-4 py-3">Balance</th>
              <th className="px-4 py-3">Total P&L</th>
              <th className="px-4 py-3">Trades</th>
              <th className="px-4 py-3">Max Leverage</th>
            </tr>
          </thead>
          <tbody>
            {filteredTraders.length === 0 ? (
              <tr>
                <td colSpan={7} className="px-4 py-8 text-center text-gray-500">
                  No traders found
                </td>
              </tr>
            ) : (
              filteredTraders.map((trader, index) => (
                <tr key={trader.id} className="border-t border-trade-border hover:bg-trade-bg/50">
                  <td className="px-4 py-3 text-gray-500">
                    {index + 1}
                    {index === 0 && <span className="ml-1">🥇</span>}
                    {index === 1 && <span className="ml-1">🥈</span>}
                    {index === 2 && <span className="ml-1">🥉</span>}
                  </td>
                  <td className="px-4 py-3">
                    <Link href={`/trader/${trader.id}`} className="font-medium hover:text-blue-400">
                      {trader.username}
                    </Link>
                  </td>
                  <td className="px-4 py-3">
                    <span className={`px-2 py-1 rounded text-xs ${getTraderTypeBadge(trader.type)}`}>
                      {trader.type}
                    </span>
                  </td>
                  <td className="px-4 py-3">${trader.balance.toLocaleString()}</td>
                  <td className={`px-4 py-3 font-medium ${trader.total_pnl >= 0 ? 'text-trade-green' : 'text-trade-red'}`}>
                    {trader.total_pnl >= 0 ? '+' : ''}${trader.total_pnl.toLocaleString()}
                  </td>
                  <td className="px-4 py-3">{trader.trade_count.toLocaleString()}</td>
                  <td className="px-4 py-3">
                    <span className={`px-2 py-1 rounded text-xs ${getLeverageBadge(trader.max_leverage_used)}`}>
                      {trader.max_leverage_used}x
                    </span>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Transparency Note */}
      <div className="mt-6 text-center text-gray-500 text-sm">
        All trader statistics are public and updated in real-time.
      </div>
    </div>
  )
}
