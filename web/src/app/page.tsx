'use client'

import { useEffect } from 'react'
import dynamic from 'next/dynamic'
import OrderBook from '@/components/OrderBook'
import TradeForm from '@/components/TradeForm'
import RecentTrades from '@/components/RecentTrades'
import PositionPanel from '@/components/PositionPanel'
import MarketStats from '@/components/MarketStats'

// Dynamic import for Chart (uses browser APIs)
const Chart = dynamic(() => import('@/components/Chart'), { ssr: false })
import { useMarketStore } from '@/store/market'
import { useUserStore } from '@/store/user'

export default function TradePage() {
  const { connectWebSocket, disconnectWebSocket, fetchAll } = useMarketStore()
  const { checkAuth, fetchUserData, isAuthenticated } = useUserStore()

  useEffect(() => {
    // Check if user is authenticated and fetch initial data
    checkAuth()
    fetchAll()

    // Connect WebSocket for real-time updates
    connectWebSocket()

    return () => {
      disconnectWebSocket()
    }
  }, [checkAuth, fetchAll, connectWebSocket, disconnectWebSocket])

  // Fetch user data when authenticated
  useEffect(() => {
    if (isAuthenticated) {
      fetchUserData()
    }
  }, [isAuthenticated, fetchUserData])

  return (
    <div className="max-w-7xl mx-auto px-2 sm:px-4 py-4 sm:py-6">
      {/* Market Stats Bar */}
      <MarketStats />

      {/* Mobile: Stack vertically, Desktop: 12-column grid */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-4 mt-4">
        {/* Order Book - Hidden on mobile, shown on tablet+ */}
        <div className="hidden md:block lg:col-span-3">
          <div className="bg-trade-card rounded-lg border border-trade-border p-4">
            <h2 className="text-sm font-medium text-gray-400 mb-4">Order Book</h2>
            <OrderBook />
          </div>
        </div>

        {/* Chart Area */}
        <div className="lg:col-span-6">
          <div className="bg-trade-card rounded-lg border border-trade-border p-3 sm:p-4">
            <h2 className="text-sm font-medium text-gray-400 mb-2">R.index Chart</h2>
            <div className="h-[250px] sm:h-[340px]">
              <Chart />
            </div>
          </div>

          {/* Trade Form */}
          <div className="mt-4">
            <TradeForm />
          </div>
        </div>

        {/* Recent Trades & Positions */}
        <div className="lg:col-span-3 space-y-4">
          {/* Order Book - Mobile only */}
          <div className="md:hidden bg-trade-card rounded-lg border border-trade-border p-4 max-h-[200px] overflow-auto">
            <h2 className="text-sm font-medium text-gray-400 mb-2">Order Book</h2>
            <OrderBook />
          </div>

          <div className="bg-trade-card rounded-lg border border-trade-border p-4 h-[200px] sm:h-[280px] overflow-auto">
            <h2 className="text-sm font-medium text-gray-400 mb-2">Recent Trades</h2>
            <RecentTrades />
          </div>

          <div className="bg-trade-card rounded-lg border border-trade-border p-4">
            <h2 className="text-sm font-medium text-gray-400 mb-2">Your Position</h2>
            <PositionPanel />
          </div>
        </div>
      </div>

      {/* Transparency Banner */}
      <div className="mt-6 bg-purple-900/30 border border-purple-700 rounded-lg p-3 sm:p-4">
        <div className="flex items-start sm:items-center">
          <span className="text-purple-400 mr-2 text-lg">&#128064;</span>
          <span className="text-purple-200 text-xs sm:text-sm">
            <strong>Radical Transparency:</strong> All positions, leverage, and trades are public.
            Everyone can see who's long, who's short, and at what leverage.
          </span>
        </div>
      </div>
    </div>
  )
}
