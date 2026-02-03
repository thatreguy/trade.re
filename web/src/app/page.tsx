'use client'

import { useEffect } from 'react'
import OrderBook from '@/components/OrderBook'
import TradeForm from '@/components/TradeForm'
import RecentTrades from '@/components/RecentTrades'
import PositionPanel from '@/components/PositionPanel'
import MarketStats from '@/components/MarketStats'
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
    <div className="max-w-7xl mx-auto px-4 py-6">
      {/* Market Stats Bar */}
      <MarketStats />

      <div className="grid grid-cols-12 gap-4 mt-4">
        {/* Order Book */}
        <div className="col-span-3">
          <div className="bg-trade-card rounded-lg border border-trade-border p-4">
            <h2 className="text-sm font-medium text-gray-400 mb-4">Order Book</h2>
            <OrderBook />
          </div>
        </div>

        {/* Chart Area (placeholder) */}
        <div className="col-span-6">
          <div className="bg-trade-card rounded-lg border border-trade-border p-4 h-96">
            <h2 className="text-sm font-medium text-gray-400 mb-4">R.index Chart</h2>
            <div className="flex items-center justify-center h-full text-gray-500">
              Chart will be rendered here using TradingView Lightweight Charts
            </div>
          </div>

          {/* Trade Form */}
          <div className="mt-4">
            <TradeForm />
          </div>
        </div>

        {/* Recent Trades & Positions */}
        <div className="col-span-3 space-y-4">
          <div className="bg-trade-card rounded-lg border border-trade-border p-4">
            <h2 className="text-sm font-medium text-gray-400 mb-4">Recent Trades</h2>
            <RecentTrades />
          </div>

          <div className="bg-trade-card rounded-lg border border-trade-border p-4">
            <h2 className="text-sm font-medium text-gray-400 mb-4">Your Position</h2>
            <PositionPanel />
          </div>
        </div>
      </div>

      {/* Transparency Banner */}
      <div className="mt-6 bg-purple-900/30 border border-purple-700 rounded-lg p-4">
        <div className="flex items-center">
          <span className="text-purple-400 mr-2">&#128064;</span>
          <span className="text-purple-200 text-sm">
            <strong>Radical Transparency:</strong> All positions, leverage, and trades are public.
            Everyone can see who's long, who's short, and at what leverage. No hidden information.
          </span>
        </div>
      </div>
    </div>
  )
}
