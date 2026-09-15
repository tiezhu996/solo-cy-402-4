import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { login as apiLogin } from '@/api/auth'
import { getMe } from '@/api/user'
import type { User } from '@/types'

interface AuthState {
  token: string
  user: User | null
  login: (username: string, password: string) => Promise<void>
  fetchMe: () => Promise<void>
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: '',
      user: null,
      async login(username: string, password: string) {
        const res: any = await apiLogin({ username, password })
        set({ token: res.data.token, user: res.data.user })
      },
      async fetchMe() {
        const res: any = await getMe()
        set({ user: res.data })
      },
      logout() {
        set({ token: '', user: null })
      },
    }),
    { name: 'cylawcase-auth' },
  ),
)
