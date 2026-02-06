// API Client for Trade.re

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

// Types
export interface Trader {
  id: string
  username: string
  type: 'human' | 'bot' | 'market_maker'
  balance: number
  total_pnl: number
  trade_count: number
  max_leverage_used: number
  created_at: string
}

export interface Position {
  trader_id: string
  instrument: string
  size: number
  entry_price: number
  leverage: number
  margin: number
  unrealized_pnl: number
  realized_pnl: number
  liquidation_price: number
  updated_at: string
}

export interface Trade {
  id: string
  instrument: string
  price: number
  size: number
  buyer_id: string
  seller_id: string
  buyer_leverage: number
  seller_leverage: number
  buyer_effect: 'open' | 'close' | 'liquidation'
  seller_effect: 'open' | 'close' | 'liquidation'
  aggressor_side: 'buy' | 'sell'
  timestamp: string
}

export interface Liquidation {
  id: string
  trader_id: string
  instrument: string
  side: 'buy' | 'sell'
  size: number
  entry_price: number
  liquidation_price: number
  mark_price: number
  leverage: number
  loss: number
  insurance_fund_hit: boolean
  timestamp: string
}

export interface OrderBookLevel {
  price: number
  size: number
  order_count: number
}

export interface OrderBook {
  instrument: string
  bids: OrderBookLevel[]
  asks: OrderBookLevel[]
  timestamp: string
}

export interface OpenInterest {
  instrument: string
  total_oi: number
  long_positions: number
  short_positions: number
  avg_long_leverage: number
  avg_short_leverage: number
  timestamp: string
}

export interface MarketStats {
  instrument: string
  last_price: number
  mark_price: number
  high_24h: number
  low_24h: number
  volume_24h: number
  open_interest: number
  insurance_fund: number
  timestamp: string
}

export interface Candle {
  timestamp: string
  open: number
  high: number
  low: number
  close: number
  volume: number
}

export interface Order {
  id: string
  trader_id: string
  instrument: string
  side: 'buy' | 'sell'
  type: 'limit' | 'market'
  price: number
  size: number
  filled_size: number
  leverage: number
  status: 'pending' | 'partial' | 'filled' | 'cancelled'
  created_at: string
}

export interface AppConfig {
  timezone: string
  max_leverage: number
  instrument: string
}

// API Client
class ApiClient {
  private token: string | null = null

  setToken(token: string) {
    this.token = token
    if (typeof window !== 'undefined') {
      localStorage.setItem('token', token)
    }
  }

  getToken(): string | null {
    if (this.token) return this.token
    if (typeof window !== 'undefined') {
      return localStorage.getItem('token')
    }
    return null
  }

  clearToken() {
    this.token = null
    if (typeof window !== 'undefined') {
      localStorage.removeItem('token')
    }
  }

  private async request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...options.headers as Record<string, string>,
    }

    const token = this.getToken()
    if (token) {
      headers['Authorization'] = `Bearer ${token}`
    }

    const response = await fetch(`${API_BASE}${path}`, {
      ...options,
      headers,
    })

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Request failed' }))
      throw new Error(error.error || `HTTP ${response.status}`)
    }

    return response.json()
  }

  // Config
  async getConfig(): Promise<AppConfig> {
    return this.request('/api/v1/config')
  }

  // Auth
  async register(username: string, password: string, type: 'human' | 'bot' = 'human') {
    return this.request<{ trader: Trader; token: string }>('/api/v1/auth/register', {
      method: 'POST',
      body: JSON.stringify({ username, password, type }),
    })
  }

  async login(username: string, password: string) {
    const result = await this.request<{ trader: Trader; token: string }>('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    })
    this.setToken(result.token)
    return result
  }

  async generateApiKey() {
    return this.request<{ api_key: string }>('/api/v1/auth/api-key', {
      method: 'POST',
    })
  }

  // Traders (Public)
  async getTraders(): Promise<Trader[]> {
    return this.request('/api/v1/traders')
  }

  async getTrader(id: string): Promise<Trader> {
    return this.request(`/api/v1/traders/${id}`)
  }

  async getTraderPositions(id: string): Promise<Position[]> {
    return this.request(`/api/v1/traders/${id}/positions`)
  }

  async getTraderTrades(id: string, limit = 50): Promise<Trade[]> {
    return this.request(`/api/v1/traders/${id}/trades?limit=${limit}`)
  }

  // Market (Public)
  async getOrderBook(): Promise<OrderBook> {
    return this.request('/api/v1/market/orderbook')
  }

  async getAllPositions(): Promise<Position[]> {
    return this.request('/api/v1/market/positions')
  }

  async getOpenInterest(): Promise<OpenInterest> {
    return this.request('/api/v1/market/oi')
  }

  async getRecentTrades(limit = 50): Promise<Trade[]> {
    return this.request(`/api/v1/market/trades?limit=${limit}`)
  }

  async getRecentLiquidations(limit = 50): Promise<Liquidation[]> {
    return this.request(`/api/v1/market/liquidations?limit=${limit}`)
  }

  async getMarketStats(): Promise<MarketStats> {
    return this.request('/api/v1/market/stats')
  }

  async getCandles(interval: '1m' | '5m' | '1h' | '1d' = '1d', limit = 100): Promise<Candle[]> {
    return this.request(`/api/v1/market/candles?interval=${interval}&limit=${limit}`)
  }

  // Trading (Authenticated)
  async placeOrder(params: {
    side: 'buy' | 'sell'
    type: 'limit' | 'market'
    price?: number
    size: number
    leverage: number
  }): Promise<{ order: Order; trades: Trade[] }> {
    return this.request('/api/v1/orders', {
      method: 'POST',
      body: JSON.stringify(params),
    })
  }

  async cancelOrder(orderId: string): Promise<void> {
    return this.request(`/api/v1/orders/${orderId}`, {
      method: 'DELETE',
    })
  }

  async closePosition(): Promise<{ order: Order; trades: Trade[] }> {
    return this.request('/api/v1/positions/close', {
      method: 'POST',
    })
  }
}

export const api = new ApiClient()
