import { create } from 'zustand'
import { api, OrderBook, Trade, Position, Liquidation, MarketStats, Trader } from '@/lib/api'
import { ws } from '@/lib/websocket'

interface MarketState {
  // Data
  orderBook: OrderBook | null
  recentTrades: Trade[]
  allPositions: Position[]
  recentLiquidations: Liquidation[]
  marketStats: MarketStats | null
  traders: Trader[]
  timezone: string

  // Loading states
  isLoading: boolean
  error: string | null

  // WebSocket connection
  isConnected: boolean

  // Actions
  fetchConfig: () => Promise<void>
  fetchOrderBook: () => Promise<void>
  fetchRecentTrades: () => Promise<void>
  fetchAllPositions: () => Promise<void>
  fetchRecentLiquidations: () => Promise<void>
  fetchMarketStats: () => Promise<void>
  fetchTraders: () => Promise<void>
  fetchAll: () => Promise<void>

  // WebSocket
  connectWebSocket: () => void
  disconnectWebSocket: () => void

  // Updates from WebSocket
  addTrade: (trade: Trade) => void
  updateOrderBook: (book: OrderBook) => void
  updatePosition: (position: Position) => void
  addLiquidation: (liquidation: Liquidation) => void
}

export const useMarketStore = create<MarketState>((set, get) => ({
  // Initial state
  orderBook: null,
  recentTrades: [],
  allPositions: [],
  recentLiquidations: [],
  marketStats: null,
  traders: [],
  timezone: 'Asia/Kolkata',
  isLoading: false,
  error: null,
  isConnected: false,

  // Fetch actions
  fetchConfig: async () => {
    try {
      const config = await api.getConfig()
      set({ timezone: config.timezone })
    } catch (e) {
      console.error('Failed to fetch config:', e)
    }
  },

  fetchOrderBook: async () => {
    try {
      const orderBook = await api.getOrderBook()
      set({ orderBook })
    } catch (e) {
      console.error('Failed to fetch order book:', e)
    }
  },

  fetchRecentTrades: async () => {
    try {
      const recentTrades = await api.getRecentTrades(50)
      set({ recentTrades })
    } catch (e) {
      console.error('Failed to fetch trades:', e)
    }
  },

  fetchAllPositions: async () => {
    try {
      const allPositions = await api.getAllPositions()
      set({ allPositions })
    } catch (e) {
      console.error('Failed to fetch positions:', e)
    }
  },

  fetchRecentLiquidations: async () => {
    try {
      const recentLiquidations = await api.getRecentLiquidations(50)
      set({ recentLiquidations })
    } catch (e) {
      console.error('Failed to fetch liquidations:', e)
    }
  },

  fetchMarketStats: async () => {
    try {
      const marketStats = await api.getMarketStats()
      set({ marketStats })
    } catch (e) {
      console.error('Failed to fetch market stats:', e)
    }
  },

  fetchTraders: async () => {
    try {
      const traders = await api.getTraders()
      set({ traders })
    } catch (e) {
      console.error('Failed to fetch traders:', e)
    }
  },

  fetchAll: async () => {
    set({ isLoading: true, error: null })
    try {
      await Promise.all([
        get().fetchConfig(),
        get().fetchOrderBook(),
        get().fetchRecentTrades(),
        get().fetchAllPositions(),
        get().fetchRecentLiquidations(),
        get().fetchMarketStats(),
        get().fetchTraders(),
      ])
    } catch (e) {
      set({ error: 'Failed to fetch market data' })
    } finally {
      set({ isLoading: false })
    }
  },

  // WebSocket
  connectWebSocket: () => {
    ws.connect()

    // Subscribe to updates
    ws.on('trade', (trade: Trade) => {
      get().addTrade(trade)
    })

    ws.on('orderbook', (book: OrderBook) => {
      get().updateOrderBook(book)
    })

    ws.on('position', (position: Position) => {
      get().updatePosition(position)
    })

    ws.on('liquidation', (liquidation: Liquidation) => {
      get().addLiquidation(liquidation)
    })

    set({ isConnected: true })
  },

  disconnectWebSocket: () => {
    ws.disconnect()
    set({ isConnected: false })
  },

  // Update functions
  addTrade: (trade: Trade) => {
    set(state => ({
      recentTrades: [trade, ...state.recentTrades.slice(0, 49)]
    }))
  },

  updateOrderBook: (book: OrderBook) => {
    set({ orderBook: book })
  },

  updatePosition: (position: Position) => {
    set(state => {
      const positions = [...state.allPositions]
      const index = positions.findIndex(
        p => p.trader_id === position.trader_id && p.instrument === position.instrument
      )

      if (index >= 0) {
        if (position.size === 0) {
          // Position closed, remove it
          positions.splice(index, 1)
        } else {
          positions[index] = position
        }
      } else if (position.size !== 0) {
        // New position
        positions.push(position)
      }

      return { allPositions: positions }
    })
  },

  addLiquidation: (liquidation: Liquidation) => {
    set(state => ({
      recentLiquidations: [liquidation, ...state.recentLiquidations.slice(0, 49)]
    }))
  },
}))
