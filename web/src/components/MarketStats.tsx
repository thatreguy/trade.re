'use client'

export default function MarketStats() {
  // Mock market data
  const stats = {
    lastPrice: 1000.75,
    change24h: 2.35,
    high24h: 1025.00,
    low24h: 975.50,
    volume24h: 1250000,
    openInterest: 85000,
    fundingRate: 0.0001,
    insuranceFund: 1000000,
  }

  const isPositive = stats.change24h >= 0

  return (
    <div className="bg-trade-card rounded-lg border border-trade-border p-4">
      <div className="flex items-center justify-between">
        {/* Price & Change */}
        <div className="flex items-center space-x-6">
          <div>
            <div className="text-2xl font-bold">
              ${stats.lastPrice.toFixed(2)}
            </div>
            <div className={`text-sm ${isPositive ? 'text-trade-green' : 'text-trade-red'}`}>
              {isPositive ? '+' : ''}{stats.change24h.toFixed(2)}%
            </div>
          </div>

          <div className="h-10 border-l border-trade-border" />

          <div className="text-sm">
            <div className="text-gray-500">24h High</div>
            <div>${stats.high24h.toFixed(2)}</div>
          </div>

          <div className="text-sm">
            <div className="text-gray-500">24h Low</div>
            <div>${stats.low24h.toFixed(2)}</div>
          </div>

          <div className="text-sm">
            <div className="text-gray-500">24h Volume</div>
            <div>${(stats.volume24h / 1000000).toFixed(2)}M</div>
          </div>
        </div>

        {/* OI & Funding */}
        <div className="flex items-center space-x-6">
          <div className="text-sm">
            <div className="text-gray-500">Open Interest</div>
            <div>${(stats.openInterest / 1000).toFixed(0)}K</div>
          </div>

          <div className="text-sm">
            <div className="text-gray-500">Funding Rate</div>
            <div className={stats.fundingRate >= 0 ? 'text-trade-green' : 'text-trade-red'}>
              {(stats.fundingRate * 100).toFixed(4)}%
            </div>
          </div>

          <div className="text-sm">
            <div className="text-gray-500">Insurance Fund</div>
            <div className="text-purple-400">${(stats.insuranceFund / 1000000).toFixed(2)}M</div>
          </div>
        </div>
      </div>
    </div>
  )
}
