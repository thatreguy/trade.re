'use client'

import { useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { api, Trader, Position, Trade } from '@/lib/api'

export default function TraderProfilePage() {
  const params = useParams()
  const [trader, setTrader] = useState<Trader | null>(null)
  const [positions, setPositions] = useState<Position[]>([])
  const [trades, setTrades] = useState<Trade[]>([])
  const [activeTab, setActiveTab] = useState<'position' | 'trades' | 'stats'>('position')
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    const fetchData = async () => {
      if (!params.id) return

      setIsLoading(true)
      try {
        const traderId = params.id as string
        const [traderData, positionsData, tradesData] = await Promise.all([
          api.getTrader(traderId),
          api.getTraderPositions(traderId),
          api.getTraderTrades(traderId, 50),
        ])
        setTrader(traderData)
        setPositions(positionsData)
        setTrades(tradesData)
      } catch (e) {
        console.error('Failed to fetch trader data:', e)
      } finally {
        setIsLoading(false)
      }
    }

    fetchData()
    const interval = setInterval(fetchData, 10000) // Refresh every 10s
    return () => clearInterval(interval)
  }, [params.id])

  // Parse trader values from API strings to numbers
  const parsedTrader = trader ? {
    ...trader,
    balance: parseFloat(trader.balance) || 0,
    total_pnl: parseFloat(trader.total_pnl) || 0,
    trade_count: trader.trade_count || 0,
    max_leverage_used: trader.max_leverage_used || 0
  } : null

  // Parse position values from API strings to numbers
  const rawPosition = positions[0] || null
  const position = rawPosition ? {
    ...rawPosition,
    size: parseFloat(rawPosition.size) || 0,
    entry_price: parseFloat(rawPosition.entry_price) || 0,
    margin: parseFloat(rawPosition.margin) || 0,
    unrealized_pnl: parseFloat(rawPosition.unrealized_pnl) || 0,
    liquidation_price: parseFloat(rawPosition.liquidation_price) || 0,
    leverage: rawPosition.leverage || 1
  } : null

  // Parse trade values from API strings to numbers
  const parsedTrades = trades.map(t => ({
    ...t,
    price: parseFloat(t.price) || 0,
    size: parseFloat(t.size) || 0
  }))

  if (isLoading) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-6">
        <div className="text-center py-12 text-gray-500">Loading...</div>
      </div>
    )
  }

  if (!parsedTrader) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-6">
        <div className="text-center py-12 text-gray-500">Trader not found</div>
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

  // Get effect badge styling
  const getEffectBadge = (effect: string) => {
    switch (effect) {
      case 'open': return { color: 'text-trade-green', label: 'OPEN' }
      case 'close': return { color: 'text-yellow-500', label: 'CLOSE' }
      case 'liquidation': return { color: 'text-trade-red', label: 'LIQUIDATED' }
      default: return { color: 'text-gray-500', label: effect }
    }
  }

  // Get OI impact based on both sides
  const getOIImpact = (buyerEffect: string, sellerEffect: string) => {
    const buyerOpening = buyerEffect === 'open'
    const sellerOpening = sellerEffect === 'open'

    if (buyerOpening && sellerOpening) {
      return { label: 'OI+', color: 'text-trade-green', desc: 'Added to OI' }
    } else if (!buyerOpening && !sellerOpening) {
      return { label: 'OI-', color: 'text-trade-red', desc: 'Reduced OI' }
    } else {
      return { label: 'OI=', color: 'text-gray-500', desc: 'No OI change' }
    }
  }

  const pnlPercent = ((parsedTrader.total_pnl / 10000) * 100).toFixed(1) // Assuming 10k starting balance

  const formatDate = (dateStr: string) => {
    try {
      return new Date(dateStr).toLocaleDateString()
    } catch {
      return dateStr
    }
  }

  return (
    <div className="max-w-7xl mx-auto px-4 py-6">
      {/* Header */}
      <div className="bg-trade-card rounded-lg border border-trade-border p-6 mb-6">
        <div className="flex items-start justify-between">
          <div>
            <div className="flex items-center space-x-3">
              <h1 className="text-2xl font-bold">{parsedTrader.username}</h1>
              <span className={`px-2 py-1 rounded text-xs ${getTraderTypeBadge(parsedTrader.type)}`}>
                {parsedTrader.type}
              </span>
            </div>
            <p className="text-gray-500 text-sm mt-1">Member since {formatDate(parsedTrader.created_at)}</p>
          </div>
          <div className="text-right">
            <div className="text-sm text-gray-500">Max Leverage Used</div>
            <div className={`text-xl font-bold px-3 py-1 rounded inline-block ${getLeverageBadge(parsedTrader.max_leverage_used)}`}>
              {parsedTrader.max_leverage_used}x
            </div>
          </div>
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-4 gap-6 mt-6">
          <div>
            <div className="text-sm text-gray-500">Balance</div>
            <div className="text-xl font-bold">${parsedTrader.balance.toLocaleString()}</div>
          </div>
          <div>
            <div className="text-sm text-gray-500">Total P&L</div>
            <div className={`text-xl font-bold ${parsedTrader.total_pnl >= 0 ? 'text-trade-green' : 'text-trade-red'}`}>
              {parsedTrader.total_pnl >= 0 ? '+' : ''}${parsedTrader.total_pnl.toLocaleString()}
              <span className="text-sm ml-1">({pnlPercent}%)</span>
            </div>
          </div>
          <div>
            <div className="text-sm text-gray-500">Total Trades</div>
            <div className="text-xl font-bold">{parsedTrader.trade_count}</div>
          </div>
          <div>
            <div className="text-sm text-gray-500">Avg Trade Size</div>
            <div className="text-xl font-bold">
              {parsedTrader.trade_count > 0 ? `$${(parsedTrader.balance / parsedTrader.trade_count * 10).toFixed(0)}` : '-'}
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
                  <span className="text-xl">{Math.abs(position.size).toFixed(3)} {position.instrument}</span>
                  <span className={`px-2 py-1 rounded text-xs ${getLeverageBadge(position.leverage)}`}>
                    {position.leverage}x
                  </span>
                </div>
                <div className={`text-xl font-bold ${position.unrealized_pnl >= 0 ? 'text-trade-green' : 'text-trade-red'}`}>
                  {position.unrealized_pnl >= 0 ? '+' : ''}${position.unrealized_pnl.toFixed(2)}
                </div>
              </div>

              <div className="grid grid-cols-4 gap-4">
                <div>
                  <div className="text-sm text-gray-500">Entry Price</div>
                  <div className="font-medium">${position.entry_price.toFixed(2)}</div>
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
                  <div className="font-medium text-trade-red">${position.liquidation_price.toFixed(2)}</div>
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
                <th className="px-4 py-3">OI Impact</th>
              </tr>
            </thead>
            <tbody>
              {parsedTrades.length === 0 ? (
                <tr>
                  <td colSpan={7} className="px-4 py-8 text-center text-gray-500">
                    No trades yet
                  </td>
                </tr>
              ) : (
                parsedTrades.map((trade) => {
                  // Determine if this trader was buyer or seller
                  const isBuyer = trade.buyer_id === parsedTrader.id
                  const side = isBuyer ? 'buy' : 'sell'
                  const leverage = isBuyer ? trade.buyer_leverage : trade.seller_leverage
                  const effect = isBuyer ? trade.buyer_effect : trade.seller_effect
                  const effectBadge = getEffectBadge(effect)
                  const oiImpact = getOIImpact(trade.buyer_effect, trade.seller_effect)

                  const formatTime = (ts: string) => {
                    try {
                      const date = new Date(ts)
                      return date.toLocaleString()
                    } catch {
                      return ts
                    }
                  }

                  return (
                    <tr key={trade.id} className="border-t border-trade-border">
                      <td className="px-4 py-3 text-gray-500 text-sm">{formatTime(trade.timestamp)}</td>
                      <td className="px-4 py-3">
                        <span className={side === 'buy' ? 'text-trade-green' : 'text-trade-red'}>
                          {side.toUpperCase()}
                        </span>
                      </td>
                      <td className="px-4 py-3">${trade.price.toFixed(2)}</td>
                      <td className="px-4 py-3">{trade.size.toFixed(3)}</td>
                      <td className="px-4 py-3">
                        <span className={`px-2 py-0.5 rounded text-xs ${getLeverageBadge(leverage)}`}>
                          {leverage}x
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <span className={effectBadge.color}>{effectBadge.label}</span>
                      </td>
                      <td className="px-4 py-3">
                        <span className={`font-medium ${oiImpact.color}`} title={oiImpact.desc}>
                          {oiImpact.label}
                        </span>
                      </td>
                    </tr>
                  )
                })
              )}
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
                <span className={`px-2 py-0.5 rounded text-xs ${getLeverageBadge(parsedTrader.max_leverage_used)}`}>
                  {parsedTrader.max_leverage_used}x
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-500">Total Trades</span>
                <span>{parsedTrader.trade_count}</span>
              </div>
            </div>
          </div>

          <div className="bg-trade-card rounded-lg border border-trade-border p-6">
            <h3 className="text-lg font-medium mb-4">Performance</h3>
            <div className="space-y-3">
              <div className="flex justify-between">
                <span className="text-gray-500">Balance</span>
                <span>${parsedTrader.balance.toLocaleString()}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-500">Total P&L</span>
                <span className={parsedTrader.total_pnl >= 0 ? 'text-trade-green' : 'text-trade-red'}>
                  {parsedTrader.total_pnl >= 0 ? '+' : ''}${parsedTrader.total_pnl.toLocaleString()}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-500">ROI</span>
                <span className={parsedTrader.total_pnl >= 0 ? 'text-trade-green' : 'text-trade-red'}>
                  {pnlPercent}%
                </span>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
