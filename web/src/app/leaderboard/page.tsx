'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'

interface Trader {
  id: string
  username: string
  type: string
  balance: number
  totalPnl: number
  tradeCount: number
  maxLeverageUsed: number
  winRate: number
}

export default function LeaderboardPage() {
  const [traders, setTraders] = useState<Trader[]>([])
  const [sortBy, setSortBy] = useState<'pnl' | 'trades' | 'winrate'>('pnl')
  const [leverageFilter, setLeverageFilter] = useState<string>('all')
  const [typeFilter, setTypeFilter] = useState<string>('all')

  // Mock data
  useEffect(() => {
    setTraders([
      { id: '1', username: 'whale_master', type: 'human', balance: 125000, totalPnl: 115000, tradeCount: 342, maxLeverageUsed: 50, winRate: 68 },
      { id: '2', username: 'algo_king', type: 'bot', balance: 98500, totalPnl: 88500, tradeCount: 1205, maxLeverageUsed: 25, winRate: 55 },
      { id: '3', username: 'degen_lord', type: 'human', balance: 75200, totalPnl: 65200, tradeCount: 89, maxLeverageUsed: 150, winRate: 72 },
      { id: '4', username: 'mm_provider', type: 'market_maker', balance: 52000, totalPnl: 42000, tradeCount: 5420, maxLeverageUsed: 10, winRate: 51 },
      { id: '5', username: 'steady_eddie', type: 'human', balance: 35000, totalPnl: 25000, tradeCount: 156, maxLeverageUsed: 5, winRate: 62 },
      { id: '6', username: 'risk_taker', type: 'human', balance: 28000, totalPnl: 18000, tradeCount: 67, maxLeverageUsed: 100, winRate: 58 },
      { id: '7', username: 'sniper_bot', type: 'bot', balance: 22000, totalPnl: 12000, tradeCount: 890, maxLeverageUsed: 75, winRate: 49 },
      { id: '8', username: 'newbie_trader', type: 'human', balance: 8500, totalPnl: -1500, tradeCount: 23, maxLeverageUsed: 20, winRate: 35 },
      { id: '9', username: 'rekt_andy', type: 'human', balance: 2100, totalPnl: -7900, tradeCount: 45, maxLeverageUsed: 150, winRate: 22 },
      { id: '10', username: 'liquidated_larry', type: 'human', balance: 0, totalPnl: -10000, tradeCount: 12, maxLeverageUsed: 150, winRate: 8 },
    ])
  }, [])

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

  const filteredTraders = traders
    .filter(t => {
      if (typeFilter !== 'all' && t.type !== typeFilter) return false
      if (leverageFilter !== 'all' && getLeverageTier(t.maxLeverageUsed) !== leverageFilter) return false
      return true
    })
    .sort((a, b) => {
      switch (sortBy) {
        case 'pnl': return b.totalPnl - a.totalPnl
        case 'trades': return b.tradeCount - a.tradeCount
        case 'winrate': return b.winRate - a.winRate
        default: return 0
      }
    })

  const totalPnlPositive = traders.filter(t => t.totalPnl > 0).length
  const totalPnlNegative = traders.filter(t => t.totalPnl < 0).length
  const avgLeverage = traders.reduce((sum, t) => sum + t.maxLeverageUsed, 0) / traders.length

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
            <option value="winrate">Win Rate</option>
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
              <th className="px-4 py-3">Win Rate</th>
              <th className="px-4 py-3">Max Leverage</th>
            </tr>
          </thead>
          <tbody>
            {filteredTraders.map((trader, index) => (
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
                <td className={`px-4 py-3 font-medium ${trader.totalPnl >= 0 ? 'text-trade-green' : 'text-trade-red'}`}>
                  {trader.totalPnl >= 0 ? '+' : ''}${trader.totalPnl.toLocaleString()}
                </td>
                <td className="px-4 py-3">{trader.tradeCount.toLocaleString()}</td>
                <td className="px-4 py-3">
                  <span className={trader.winRate >= 50 ? 'text-trade-green' : 'text-trade-red'}>
                    {trader.winRate}%
                  </span>
                </td>
                <td className="px-4 py-3">
                  <span className={`px-2 py-1 rounded text-xs ${getLeverageBadge(trader.maxLeverageUsed)}`}>
                    {trader.maxLeverageUsed}x
                  </span>
                </td>
              </tr>
            ))}
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
