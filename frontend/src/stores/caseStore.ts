import { create } from 'zustand'
import { listCases } from '@/api/case'
import type { CaseItem } from '@/types'

interface CaseState {
  list: CaseItem[]
  total: number
  fetchList: (params?: Record<string, unknown>) => Promise<void>
}

export const useCaseStore = create<CaseState>((set) => ({
  list: [],
  total: 0,
  async fetchList(params = {}) {
    const res: any = await listCases(params)
    set({ list: res.data.list, total: res.data.total })
  },
}))
