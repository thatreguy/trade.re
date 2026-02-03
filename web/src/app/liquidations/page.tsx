'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'

interface Liquidation {
  id: string
  traderId: string
  username: string
  traderType: string
  side: 'long' | 'short'
  size: number
  entryPrice: number
  liquidationPrice: number
  markPrice: number
  leverage: number
  loss: number
  insuranceFundHit: boolean
  timestamp: string
}

export default function LiquidationsPage() {
  const [liquidations, setLiquidations] = useState<Liquidation[]>([])
  const [sideFilter, setSideFilter] = useState<'all' | 'long' | 'short'>('all')
  const [insuranceFund, setInsuranceFund] = useState(985000)

  // Mock data
  useEffect(() => {
    setLiquidations([
      { id: '1', traderId: '9', username: 'rekt_andy', traderType: 'human', side: 'long', size: 50, entryPrice: 1010, liquidationPrice: 1003.37, markPrice: 1002.50, leverage: 150, loss: 375, insuranceFundHit: false, timestamp: '2 min ago' },
      { id: '2', traderId: '10', username: 'liquidated_larry', traderType: 'human', side: 'short', size: 30, entryPrice: 990, liquidationPrice: 996.60, markPrice: 997.00, leverage: 150, loss: 210, insuranceFundHit: true, timestamp: '15 min ago' },
      { id: '3', traderId: '8', username: 'newbie_trader', traderType: 'human', side: 'long', size: 10, entryPrice: 1005, liquidationPrice: 995.05, markPrice: 994.00, leverage: 100, loss: 110, insuranceFundHit: false, timestamp: '1 hour ago' },
      { id: '4', traderId: '7', username: 'sniper_bot', traderType: 'bot', side: 'short', size: 25, entryPrice: 985, liquidationPrice: 1004.70, markPrice: 1005.00, leverage: 50, loss: 500, insuranceFundHit: false, timestamp: '2 hours ago' },
      { id: '5', traderId: '6', username: 'risk_taker', traderType: 'human', side: 'long', size: 100, entryPrice: 1020, liquidationPrice: 1009.80, markPrice: 1008.00, leverage: 100, loss: 1200, insuranceFundHit: true, timestamp: '5 hours ago' },
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

  const filteredLiquidations = liquidations.filter(l => {
    if (sideFilter !== 'all' && l.side !== sideFilter) return false
    return true
  })

  const totalLiquidations = liquidations.length
  const totalLongLiqs = liquidations.filter(l => l.side === 'long').length
  const totalShortLiqs = liquidations.filter(l => l.side === 'short').length
  const totalLoss = liquidations.reduce((sum, l) => sum + l.loss, 0)
  const avgLeverageLiquidated = liquidations.reduce((sum, l) => sum + l.leverage, 0) / liquidations.length

  return (
    <div className="max-w-7xl mx-auto px-4 py-6">
      <h1 className="text-2xl font-bold mb-2">Liquidations</h1>
      <p className="text-gray-500 mb-6">Real-time liquidation feed. See who got rekt and at what leverage.</p>

      {/* Summary Stats */}
      <div className="grid grid-cols-5 gap-4 mb-6">
        <div className="bg-trade-card rounded-lg border border-trade-border p-4">
          <div className="text-gray-500 text-sm">Total Liquidations</div>
          <div className="text-xl font-bold text-trade-red">{totalLiquidations}</div>
        </div>
        <div className="bg-trade-card rounded-lg border border-trade-border p-4">
          <div className="text-gray-500 text-sm">Longs Liquidated</div>
          <div className="text-xl font-bold text-trade-green">{totalLongLiqs}</div>
        </div>
        <div className="bg-trade-card rounded-lg border border-trade-border p-4">
          <div className="text-gray-500 text-sm">Shorts Liquidated</div>
          <div className="text-xl font-bold text-trade-red">{totalShortLiqs}</div>
        </div>
        <div className="bg-trade-card rounded-lg border border-trade-border p-4">
          <div className="text-gray-500 text-sm">Total Loss</div>
          <div className="text-xl font-bold">${totalLoss.toLocaleString()}</div>
        </div>
        <div className="bg-trade-card rounded-lg border border-trade-border p-4">
          <div className="text-gray-500 text-sm">Insurance Fund</div>
          <div className="text-xl font-bold text-purple-400">${insuranceFund.toLocaleString()}</div>
        </div>
      </div>

      {/* Leverage Distribution of Liquidations */}
      <div className="bg-trade-card rounded-lg border border-trade-border p-4 mb-6">
        <h3 className="text-sm font-medium text-gray-400 mb-3">Liquidations by Leverage Tier</h3>
        <div className="flex space-x-4">
          <div className="flex-1">
            <div className="text-xs text-gray-500 mb-1">Conservative (1-10x)</div>
            <div className="h-2 bg-trade-bg rounded overflow-hidden">
              <div className="h-full leverage-conservative" style={{ width: '5%' }}></div>
            </div>
            <div className="text-xs text-gray-500 mt-1">{liquidations.filter(l => l.leverage <= 10).length}</div>
          </div>
          <div className="flex-1">
            <div className="text-xs text-gray-500 mb-1">Moderate (11-50x)</div>
            <div className="h-2 bg-trade-bg rounded overflow-hidden">
              <div className="h-full leverage-moderate" style={{ width: '20%' }}></div>
            </div>
            <div className="text-xs text-gray-500 mt-1">{liquidations.filter(l => l.leverage > 10 && l.leverage <= 50).length}</div>
          </div>
          <div className="flex-1">
            <div className="text-xs text-gray-500 mb-1">Aggressive (51-100x)</div>
            <div className="h-2 bg-trade-bg rounded overflow-hidden">
              <div className="h-full leverage-aggressive" style={{ width: '35%' }}></div>
            </div>
            <div className="text-xs text-gray-500 mt-1">{liquidations.filter(l => l.leverage > 50 && l.leverage <= 100).length}</div>
          </div>
          <div className="flex-1">
            <div className="text-xs text-gray-500 mb-1">Degen (101-150x)</div>
            <div className="h-2 bg-trade-bg rounded overflow-hidden">
              <div className="h-full leverage-degen" style={{ width: '40%' }}></div>
            </div>
            <div className="text-xs text-gray-500 mt-1">{liquidations.filter(l => l.leverage > 100).length}</div>
          </div>
        </div>
        <div className="mt-3 text-xs text-gray-500">
          Average leverage at liquidation: <span className="font-bold">{avgLeverageLiquidated.toFixed(0)}x</span>
        </div>
      </div>

      {/* Filters */}
      <div className="flex space-x-4 mb-4">
        <button
          className={`px-4 py-2 rounded ${sideFilter === 'all' ? 'bg-white text-black' : 'bg-trade-card text-gray-400'}`}
          onClick={() => setSideFilter('all')}
        >
          All
        </button>
        <button
          className={`px-4 py-2 rounded ${sideFilter === 'long' ? 'bg-trade-green text-white' : 'bg-trade-card text-gray-400'}`}
          onClick={() => setSideFilter('long')}
        >
          Longs
        </button>
        <button
          className={`px-4 py-2 rounded ${sideFilter === 'short' ? 'bg-trade-red text-white' : 'bg-trade-card text-gray-400'}`}
          onClick={() => setSideFilter('short')}
        >
          Shorts
        </button>
      </div>

      {/* Liquidations Feed */}
      <div className="space-y-3">
        {filteredLiquidations.map((liq) => (
          <div key={liq.id} className="bg-trade-card rounded-lg border border-trade-border p-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-4">
                <div className="text-2xl">💀</div>
                <div>
                  <div className="flex items-center space-x-2">
                    <Link href={`/trader/${liq.traderId}`} className="font-medium hover:text-blue-400">
                      {liq.username}
                    </Link>
                    <span className={`px-2 py-0.5 rounded text-xs ${getTraderTypeBadge(liq.traderType)}`}>
                      {liq.traderType}
                    </span>
                    <span className={`px-2 py-0.5 rounded text-xs ${getLeverageBadge(liq.leverage)}`}>
                      {liq.leverage}x
                    </span>
                  </div>
                  <div className="text-sm text-gray-500">
                    <span className={liq.side === 'long' ? 'text-trade-green' : 'text-trade-red'}>
                      {liq.side.toUpperCase()}
                    </span>
                    {' '}{liq.size} @ ${liq.entryPrice.toFixed(2)} → liquidated @ ${liq.markPrice.toFixed(2)}
                  </div>
                </div>
              </div>
              <div className="text-right">
                <div className="text-trade-red font-bold">-${liq.loss.toFixed(2)}</div>
                <div className="text-xs text-gray-500">{liq.timestamp}</div>
                {liq.insuranceFundHit && (
                  <div className="text-xs text-purple-400">Insurance fund hit</div>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>

      {filteredLiquidations.length === 0 && (
        <div className="text-center py-12 text-gray-500">
          No liquidations matching your filter
        </div>
      )}

      {/* Info Box */}
      <div className="mt-6 bg-red-900/20 border border-red-700 rounded-lg p-4">
        <h3 className="font-medium text-red-400 mb-2">How Liquidations Work</h3>
        <ul className="text-sm text-gray-400 space-y-1">
          <li>• Positions are liquidated when mark price crosses the liquidation price</li>
          <li>• Higher leverage = closer liquidation price = higher risk</li>
          <li>• Margin covers the loss; excess loss hits the insurance fund</li>
          <li>• All liquidations are public - everyone sees who got rekt</li>
        </ul>
      </div>
    </div>
  )
}
