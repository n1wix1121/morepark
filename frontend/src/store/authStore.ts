import { create } from 'zustand'
import api from '@/api/client'

interface User {
  id: string
  email: string
  full_name: string
  role: 'director' | 'cashier' | 'lifeguard' | 'technician' | 'barman'
  is_active: boolean
}

interface AuthState {
  token: string | null
  user: User | null
  isLoading: boolean
  login: (email: string, password: string) => Promise<void>
  logout: () => void
  fetchMe: () => Promise<void>
}

export const useAuthStore = create<AuthState>((set) => ({
  token: localStorage.getItem('token'),
  user: JSON.parse(localStorage.getItem('user') || 'null'),
  isLoading: false,

  login: async (email, password) => {
    set({ isLoading: true })
    try {
      const res = await api.post('/auth/login', { email, password })
      const { token, user } = res.data

      localStorage.setItem('token', token)
      localStorage.setItem('user', JSON.stringify(user))

      set({ token, user, isLoading: false })
    } catch (error) {
      set({ isLoading: false })
      throw error
    }
  },

  logout: () => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    set({ token: null, user: null })
  },

  fetchMe: async () => {
    try {
      const res = await api.get('/auth/me')
      set({ user: res.data.user })
    } catch (error) {
      console.error('Failed to fetch user:', error)
    }
  },
}))