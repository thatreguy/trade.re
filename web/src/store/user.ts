import { create } from 'zustand'
import { api, Trader, Position, Trade } from '@/lib/api'

interface UserState {
  // Current user
  user: Trader | null
  isAuthenticated: boolean

  // User's data
  position: Position | null
  trades: Trade[]

  // Loading
  isLoading: boolean
  error: string | null

  // Actions
  login: (username: string, password: string) => Promise<boolean>
  register: (username: string, password: string, type?: 'human' | 'bot') => Promise<boolean>
  logout: () => void
  checkAuth: () => Promise<void>
  fetchUserData: () => Promise<void>
}

export const useUserStore = create<UserState>((set, get) => ({
  user: null,
  isAuthenticated: false,
  position: null,
  trades: [],
  isLoading: false,
  error: null,

  login: async (username: string, password: string) => {
    set({ isLoading: true, error: null })
    try {
      const result = await api.login(username, password)
      set({
        user: result.trader,
        isAuthenticated: true,
        isLoading: false,
      })
      await get().fetchUserData()
      return true
    } catch (e) {
      set({
        error: e instanceof Error ? e.message : 'Login failed',
        isLoading: false,
      })
      return false
    }
  },

  register: async (username: string, password: string, type: 'human' | 'bot' = 'human') => {
    set({ isLoading: true, error: null })
    try {
      const result = await api.register(username, password, type)
      api.setToken(result.token)
      set({
        user: result.trader,
        isAuthenticated: true,
        isLoading: false,
      })
      return true
    } catch (e) {
      set({
        error: e instanceof Error ? e.message : 'Registration failed',
        isLoading: false,
      })
      return false
    }
  },

  logout: () => {
    api.clearToken()
    set({
      user: null,
      isAuthenticated: false,
      position: null,
      trades: [],
    })
  },

  checkAuth: async () => {
    const token = api.getToken()
    if (!token) {
      set({ isAuthenticated: false, user: null })
      return
    }

    // Token exists, try to validate by fetching user data
    // For now, just mark as authenticated
    // TODO: Add a /me endpoint to validate token
    set({ isAuthenticated: true })
  },

  fetchUserData: async () => {
    const { user } = get()
    if (!user) return

    try {
      const [positions, trades] = await Promise.all([
        api.getTraderPositions(user.id),
        api.getTraderTrades(user.id, 20),
      ])

      set({
        position: positions[0] || null, // R.index position
        trades,
      })
    } catch (e) {
      console.error('Failed to fetch user data:', e)
    }
  },
}))
