'use client'

import { useState } from 'react'
import { useUserStore } from '@/store/user'
import { useMarketStore } from '@/store/market'
import { api } from '@/lib/api'

export default function TradeForm() {
  const { isAuthenticated, fetchUserData } = useUserStore()
  const { marketStats } = useMarketStore()

  const [side, setSide] = useState<'buy' | 'sell'>('buy')
  const [orderType, setOrderType] = useState<'limit' | 'market'>('limit')
  const lastPrice = parseFloat(marketStats?.last_price) || 1000
  const [price, setPrice] = useState(lastPrice.toFixed(2))
  const [size, setSize] = useState('1.0')
  const [leverage, setLeverage] = useState(10)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const leverageTier = leverage <= 10 ? 'conservative' : leverage <= 50 ? 'moderate' : leverage <= 100 ? 'aggressive' : 'degen'

  const handleSubmit = async () => {
    if (!isAuthenticated) {
      setError('Please login to trade')
      return
    }

    setIsSubmitting(true)
    setError(null)

    try {
      await api.placeOrder({
        side,
        type: orderType,
        price: orderType === 'limit' ? parseFloat(price) : undefined,
        size: parseFloat(size),
        leverage,
      })
      await fetchUserData()
      setSize('1.0') // Reset size after order
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Order failed')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="bg-trade-card rounded-lg border border-trade-border p-4">
      {/* Buy/Sell Toggle */}
      <div className="flex mb-4">
        <button
          className={`flex-1 py-2 rounded-l font-medium ${
            side === 'buy' ? 'bg-trade-green text-white' : 'bg-trade-border text-gray-400'
          }`}
          onClick={() => setSide('buy')}
        >
          Long
        </button>
        <button
          className={`flex-1 py-2 rounded-r font-medium ${
            side === 'sell' ? 'bg-trade-red text-white' : 'bg-trade-border text-gray-400'
          }`}
          onClick={() => setSide('sell')}
        >
          Short
        </button>
      </div>

      {/* Order Type */}
      <div className="flex mb-4 text-sm">
        <button
          className={`px-4 py-1 ${orderType === 'limit' ? 'text-white border-b-2 border-white' : 'text-gray-500'}`}
          onClick={() => setOrderType('limit')}
        >
          Limit
        </button>
        <button
          className={`px-4 py-1 ${orderType === 'market' ? 'text-white border-b-2 border-white' : 'text-gray-500'}`}
          onClick={() => setOrderType('market')}
        >
          Market
        </button>
      </div>

      {/* Price Input */}
      {orderType === 'limit' && (
        <div className="mb-4">
          <label className="text-xs text-gray-500 block mb-1">Price</label>
          <input
            type="text"
            value={price}
            onChange={(e) => setPrice(e.target.value)}
            className="w-full bg-trade-bg border border-trade-border rounded px-3 py-2 text-white"
          />
        </div>
      )}

      {/* Size Input */}
      <div className="mb-4">
        <label className="text-xs text-gray-500 block mb-1">Size</label>
        <input
          type="text"
          value={size}
          onChange={(e) => setSize(e.target.value)}
          className="w-full bg-trade-bg border border-trade-border rounded px-3 py-2 text-white"
        />
      </div>

      {/* Leverage Slider */}
      <div className="mb-4">
        <div className="flex justify-between items-center mb-2">
          <label className="text-xs text-gray-500">Leverage</label>
          <span className={`text-sm font-bold px-2 py-0.5 rounded leverage-${leverageTier}`}>
            {leverage}x
          </span>
        </div>
        <input
          type="range"
          min="1"
          max="150"
          value={leverage}
          onChange={(e) => setLeverage(parseInt(e.target.value))}
          className="w-full"
        />
        <div className="flex justify-between text-xs text-gray-500 mt-1">
          <span>1x</span>
          <span>10x</span>
          <span>50x</span>
          <span>100x</span>
          <span>150x</span>
        </div>
      </div>

      {/* Transparency Notice */}
      <div className="mb-4 p-2 bg-yellow-900/30 border border-yellow-700 rounded text-xs text-yellow-200">
        &#128064; Your leverage ({leverage}x) will be publicly visible to all traders
      </div>

      {/* Order Summary */}
      <div className="mb-4 p-3 bg-trade-bg rounded text-sm">
        <div className="flex justify-between mb-1">
          <span className="text-gray-500">Notional</span>
          <span>${(parseFloat(size) * parseFloat(price)).toFixed(2)}</span>
        </div>
        <div className="flex justify-between mb-1">
          <span className="text-gray-500">Margin Required</span>
          <span>${(parseFloat(size) * parseFloat(price) / leverage).toFixed(2)}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-gray-500">Liquidation Price</span>
          <span className="text-trade-red">
            ${side === 'buy'
              ? (parseFloat(price) * (1 - 1 / leverage * 0.95)).toFixed(2)
              : (parseFloat(price) * (1 + 1 / leverage * 0.95)).toFixed(2)
            }
          </span>
        </div>
      </div>

      {/* Error Display */}
      {error && (
        <div className="mb-4 p-2 bg-red-900/30 border border-red-700 rounded text-xs text-red-300">
          {error}
        </div>
      )}

      {/* Submit Button */}
      <button
        onClick={handleSubmit}
        disabled={isSubmitting || !isAuthenticated}
        className={`w-full py-3 rounded font-bold ${
          side === 'buy' ? 'bg-trade-green hover:bg-green-600' : 'bg-trade-red hover:bg-red-600'
        } text-white disabled:opacity-50 disabled:cursor-not-allowed`}
      >
        {!isAuthenticated
          ? 'Login to Trade'
          : isSubmitting
          ? 'Submitting...'
          : `${side === 'buy' ? 'Long' : 'Short'} R.index`}
      </button>
    </div>
  )
}
