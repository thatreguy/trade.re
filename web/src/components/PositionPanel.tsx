'use client'

export default function PositionPanel() {
  // Mock position data
  const position = {
    size: 5.0,
    entryPrice: 998.50,
    leverage: 25,
    margin: 199.70,
    unrealizedPnl: 12.50,
    liquidationPrice: 960.12,
    markPrice: 1001.00,
  }

  const isLong = position.size > 0
  const pnlPercent = ((position.unrealizedPnl / position.margin) * 100).toFixed(2)
  const leverageTier = position.leverage <= 10 ? 'conservative' : position.leverage <= 50 ? 'moderate' : position.leverage <= 100 ? 'aggressive' : 'degen'

  if (!position.size) {
    return (
      <div className="text-gray-500 text-sm text-center py-4">
        No open position
      </div>
    )
  }

  return (
    <div className="text-sm">
      {/* Position Header */}
      <div className="flex items-center justify-between mb-3">
        <span className={`font-bold ${isLong ? 'text-trade-green' : 'text-trade-red'}`}>
          {isLong ? 'LONG' : 'SHORT'} {Math.abs(position.size)}
        </span>
        <span className={`px-2 py-0.5 rounded text-xs leverage-${leverageTier}`}>
          {position.leverage}x
        </span>
      </div>

      {/* Position Details */}
      <div className="space-y-2">
        <div className="flex justify-between">
          <span className="text-gray-500">Entry Price</span>
          <span>${position.entryPrice.toFixed(2)}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-gray-500">Mark Price</span>
          <span>${position.markPrice.toFixed(2)}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-gray-500">Margin</span>
          <span>${position.margin.toFixed(2)}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-gray-500">Unrealized P&L</span>
          <span className={position.unrealizedPnl >= 0 ? 'text-trade-green' : 'text-trade-red'}>
            {position.unrealizedPnl >= 0 ? '+' : ''}${position.unrealizedPnl.toFixed(2)} ({pnlPercent}%)
          </span>
        </div>
        <div className="flex justify-between">
          <span className="text-gray-500">Liq. Price</span>
          <span className="text-trade-red">${position.liquidationPrice.toFixed(2)}</span>
        </div>
      </div>

      {/* Close Button */}
      <button className="w-full mt-4 py-2 bg-trade-border hover:bg-gray-600 rounded text-sm">
        Close Position
      </button>
    </div>
  )
}
