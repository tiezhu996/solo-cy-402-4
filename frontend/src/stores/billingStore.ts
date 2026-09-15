import { create } from 'zustand'
import { listBillings, getBillingSummary, listBillingsByCase } from '@/api/billing'
import type { Billing } from '@/types'

interface BillingState {
  list: Billing[]
  total: number
  byCase: Billing[]
  summary: { receivable: number; received: number; pending: number }
  fetchList: (params?: Record<string, unknown>) => Promise<void>
  fetchSummary: () => Promise<void>
  fetchByCase: (caseId: number) => Promise<void>
}

export const useBillingStore = create<BillingState>((set) => ({
  list: [],
  total: 0,
  byCase: [],
  summary: { receivable: 0, received: 0, pending: 0 },
  async fetchList(params = {}) {
    const res: any = await listBillings(params)
    set({ list: res.data.list, total: res.data.total })
  },
  async fetchSummary() {
    const res: any = await getBillingSummary()
    set({ summary: res.data })
  },
  async fetchByCase(caseId: number) {
    const res: any = await listBillingsByCase(caseId)
    set({ byCase: res.data })
  },
}))
