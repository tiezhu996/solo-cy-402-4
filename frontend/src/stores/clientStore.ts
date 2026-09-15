import { create } from 'zustand'
import { listClients } from '@/api/client'
import type { Client } from '@/types'

interface ClientState {
  list: Client[]
  total: number
  fetchList: (params?: Record<string, unknown>) => Promise<void>
}

export const useClientStore = create<ClientState>((set) => ({
  list: [],
  total: 0,
  async fetchList(params = {}) {
    const res: any = await listClients(params)
    set({ list: res.data.list, total: res.data.total })
  },
}))
