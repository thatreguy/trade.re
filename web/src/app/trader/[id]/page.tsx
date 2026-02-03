'use client'

import { useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'

interface Trader {
  id: string
  username: string
  type: string
  balance: number
  totalPnl: number
  tradeCount: number
  maxLeverageUsed: number
  createdAt: string
}

interface Position {
  instrument: string
  size: number
  entryPrice: number
  leverage: number
  margin: number
  unrealizedPnl: number
  liquidationPrice: number
}

interface Trade {
  id: string
  timestamp: string
  price: number
  size: number
  side: 'buy' | 'sell'
  leverage: number
  effect: string
  pnl: number
}

export default function TraderProfilePage() {
  const params = useParams()
  const [trader, setTrader] = useState<Trader | null>(null)
  const [position, setPosition] = useState<Position | null>(null)
  const [trades, setTrades] = useState<Trade[]>([])
  const [activeTab, setActiveTab] = useState<'position' | 'trades' | 'stats'>('position')

  // Mock data
  useEffect(() => {
    // Simulating fetching trader data
    setTrader({
      id: params.id as string,
      username: 'whale_master',
      type: 'human',
      balance: 125000,
      totalPnl: 115000,
      tradeCount: 342,
      maxLeverageUsed: 50,
      createdAt: '2024-01-15'
    })

    setPosition({
      instrument: 'R.index',
      size: 50,
      entryPrice: 995.00,
      leverage: 25,
      margin: 1990,
      unrealizedPnl: 275,
      liquidationPrice: 955.20
    })

    setTrades([
      { id: '1', timestamp: '10 min ago', price: 1000.50, size: 10, side: 'buy', leverage: 25, effect: 'open', pnl: 0 },
      { id: '2', timestamp: '1 hour ago', price: 998.00, size: 15, side: 'buy', leverage: 25, effect: 'open', pnl: 0 },
      { id: '3', timestamp: '2 hours ago', price: 1002.00, size: 5, side: 'sell', leverage: 50, effect: 'close', pnl: 25 },
      { id: '4', timestamp: '3 hours ago', price: 995.00, size: 20, side: 'buy', leverage: 50, effect: 'open', pnl: 0 },
      { id: '5', timestamp: '5 hours ago', price: 990.00, size: 30, side: 'sell', leverage: 10, effect: 'close', pnl: 150 },
    ])
  }, [params.id])

  if (!trader) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-6">
        <div className="text-center py-12 text-gray-500">Loading...</div>
      </div>
    )
  }

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

  const pnlPercent = ((trader.totalPnl / 10000) * 100).toFixed(1) // Assuming 10k starting balance

  return (
    <div className="max-w-7xl mx-auto px-4 py-6">
      {/* Header */}
      <div className="bg-trade-card rounded-lg border border-trade-border p-6 mb-6">
        <div className="flex items-start justify-between">
          <div>
            <div className="flex items-center space-x-3">
              <h1 className="text-2xl font-bold">{trader.username}</h1>
              <span className={`px-2 py-1 rounded text-xs ${getTraderTypeBadge(trader.type)}`}>
                {trader.type}
              </span>
            </div>
            <p className="text-gray-500 text-sm mt-1">Member since {trader.createdAt}</p>
          </div>
          <div className="text-right">
            <div className="text-sm text-gray-500">Max Leverage Used</div>
            <div className={`text-xl font-bold px-3 py-1 rounded inline-block ${getLeverageBadge(trader.maxLeverageUsed)}`}>
              {trader.maxLeverageUsed}x
            </div>
          </div>
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-4 gap-6 mt-6">
          <div>
            <div className="text-sm text-gray-500">Balance</div>
            <div className="text-xl font-bold">${trader.balance.toLocaleString()}</div>
          </div>
          <div>
            <div className="text-sm text-gray-500">Total P&L</div>
            <div className={`text-xl font-bold ${trader.totalPnl >= 0 ? 'text-trade-green' : 'text-trade-red'}`}>
              {trader.totalPnl >= 0 ? '+' : ''}${trader.totalPnl.toLocaleString()}
              <span className="text-sm ml-1">({pnlPercent}%)</span>
            </div>
          </div>
          <div>
            <div className="text-sm text-gray-500">Total Trades</div>
            <div className="text-xl font-bold">{trader.tradeCount}</div>
          </div>
          <div>
            <div className="text-sm text-gray-500">Avg Trade Size</div>
            <div className="text-xl font-bold">
              ${(trader.balance / trader.tradeCount * 10).toFixed(0)}
            </div>
          </div>
        </div>
      </div>

      {/* Transparency Banner */}
      <div className="bg-purple-900/30 border border-purple-700 rounded-lg p-3 mb-6">
        <div className="flex items-center">
          <span className="text-purple-400 mr-2">&#128064;</span>
          <span className="text-purple-200 text-sm">
            All data on this page is public. Everyone can see this trader's positions, trades, and leverage.
          </span>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex space-x-1 mb-6 border-b border-trade-border">
        <button
          className={`px-4 py-2 -mb-px ${activeTab === 'position' ? 'border-b-2 border-white text-white' : 'text-gray-500'}`}
          onClick={() => setActiveTab('position')}
        >
          Current Position
        </button>
        <button
          className={`px-4 py-2 -mb-px ${activeTab === 'trades' ? 'border-b-2 border-white text-white' : 'text-gray-500'}`}
          onClick={() => setActiveTab('trades')}
        >
          Trade History
        </button>
        <button
          className={`px-4 py-2 -mb-px ${activeTab === 'stats' ? 'border-b-2 border-white text-white' : 'text-gray-500'}`}
          onClick={() => setActiveTab('stats')}
        >
          Statistics
        </button>
      </div>

      {/* Position Tab */}
      {activeTab === 'position' && (
        <div className="bg-trade-card rounded-lg border border-trade-border p-6">
          {position && position.size !== 0 ? (
            <div>
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center space-x-3">
                  <span className={`text-xl font-bold ${position.size > 0 ? 'text-trade-green' : 'text-trade-red'}`}>
                    {position.size > 0 ? 'LONG' : 'SHORT'}
                  </span>
                  <span className="text-xl">{Math.abs(position.size)} {position.instrument}</span>
                  <span className={`px-2 py-1 rounded text-xs ${getLeverageBadge(position.leverage)}`}>
                    {position.leverage}x
                  </span>
                </div>
                <div className={`text-xl font-bold ${position.unrealizedPnl >= 0 ? 'text-trade-green' : 'text-trade-red'}`}>
                  {position.unrealizedPnl >= 0 ? '+' : ''}${position.unrealizedPnl.toFixed(2)}
                </div>
              </div>

              <div className="grid grid-cols-4 gap-4">
                <div>
                  <div className="text-sm text-gray-500">Entry Price</div>
                  <div className="font-medium">${position.entryPrice.toFixed(2)}</div>
                </div>
                <div>
                  <div className="text-sm text-gray-500">Margin</div>
                  <div className="font-medium">${position.margin.toFixed(2)}</div>
                </div>
                <div>
                  <div className="text-sm text-gray-500">Leverage</div>
                  <div className="font-medium">{position.leverage}x</div>
                </div>
                <div>
                  <div className="text-sm text-gray-500">Liquidation Price</div>
                  <div className="font-medium text-trade-red">${position.liquidationPrice.toFixed(2)}</div>
                </div>
              </div>
            </div>
          ) : (
            <div className="text-center py-8 text-gray-500">
              No open position
            </div>
          )}
        </div>
      )}

      {/* Trades Tab */}
      {activeTab === 'trades' && (
        <div className="bg-trade-card rounded-lg border border-trade-border overflow-hidden">
          <table className="w-full">
            <thead className="bg-trade-bg">
              <tr className="text-left text-sm text-gray-500">
                <th className="px-4 py-3">Time</th>
                <th className="px-4 py-3">Side</th>
                <th className="px-4 py-3">Price</th>
                <th className="px-4 py-3">Size</th>
                <th className="px-4 py-3">Leverage</th>
                <th className="px-4 py-3">Effect</th>
                <th className="px-4 py-3">P&L</th>
              </tr>
            </thead>
            <tbody>
              {trades.map((trade) => (
                <tr key={trade.id} className="border-t border-trade-border">
                  <td className="px-4 py-3 text-gray-500">{trade.timestamp}</td>
                  <td className="px-4 py-3">
                    <span className={trade.side === 'buy' ? 'text-trade-green' : 'text-trade-red'}>
                      {trade.side.toUpperCase()}
                    </span>
                  </td>
                  <td className="px-4 py-3">${trade.price.toFixed(2)}</td>
                  <td className="px-4 py-3">{trade.size}</td>
                  <td className="px-4 py-3">
                    <span className={`px-2 py-0.5 rounded text-xs ${getLeverageBadge(trade.leverage)}`}>
                      {trade.leverage}x
                    </span>
                  </td>
                  <td className="px-4 py-3 text-gray-500">{trade.effect}</td>
                  <td className={`px-4 py-3 ${trade.pnl > 0 ? 'text-trade-green' : trade.pnl < 0 ? 'text-trade-red' : 'text-gray-500'}`}>
                    {trade.pnl !== 0 ? (trade.pnl > 0 ? '+' : '') + '$' + trade.pnl.toFixed(2) : '-'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Stats Tab */}
      {activeTab === 'stats' && (
        <div className="grid grid-cols-2 gap-6">
          <div className="bg-trade-card rounded-lg border border-trade-border p-6">
            <h3 className="text-lg font-medium mb-4">Leverage Usage</h3>
            <div className="space-y-3">
              <div className="flex justify-between">
                <span className="text-gray-500">Max Leverage Used</span>
                <span className={`px-2 py-0.5 rounded text-xs ${getLeverageBadge(trader.maxLeverageUsed)}`}>
                  {trader.maxLeverageUsed}x
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-500">Avg Leverage</span>
                <span>25x</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-500">Trades at 100x+</span>
                <span>12</span>
              </div>
            </div>
          </div>

          <div className="bg-trade-card rounded-lg border border-trade-border p-6">
            <h3 className="text-lg font-medium mb-4">Performance</h3>
            <div className="space-y-3">
              <div className="flex justify-between">
                <span className="text-gray-500">Win Rate</span>
                <span className="text-trade-green">68%</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-500">Profit Factor</span>
                <span>2.4</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-500">Liquidations</span>
                <span className="text-trade-red">3</span>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
