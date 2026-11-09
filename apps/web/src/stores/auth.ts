import { defineStore } from 'pinia'
import { api, clearSession, getToken, loadUser, saveSession } from '../api'
import type { LoginResponse, MeResponse, User } from '../types'

interface AuthState {
  token: string
  user: User | null
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    token: getToken(),
    user: loadUser(),
  }),
  getters: {
    isLoggedIn: (state): boolean => !!state.token,
    isAdmin: (state): boolean => !!state.user?.is_admin,
  },
  actions: {
    async login(username: string, password: string): Promise<void> {
      const data = await api<LoginResponse>('/auth/login', {
        method: 'POST',
        body: { username, password },
      })
      this.setSession(data.token, data.user)
    },
    async autologin(key: string): Promise<void> {
      const data = await api<LoginResponse>(`/auth/autologin?key=${encodeURIComponent(key)}`)
      this.setSession(data.token, data.user)
    },
    async fetchMe(): Promise<void> {
      const data = await api<MeResponse>('/auth/me')
      this.user = data.user
      saveSession(this.token, data.user)
    },
    setSession(token: string, user: User): void {
      this.token = token
      this.user = user
      saveSession(token, user)
    },
    logout(): void {
      this.token = ''
      this.user = null
      clearSession()
    },
  },
})
