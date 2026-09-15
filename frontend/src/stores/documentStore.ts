import { create } from 'zustand'
import { listDocuments, listDocumentsByCase } from '@/api/document'
import type { DocumentItem } from '@/types'

interface DocumentState {
  list: DocumentItem[]
  total: number
  byCase: DocumentItem[]
  fetchList: (params?: Record<string, unknown>) => Promise<void>
  fetchByCase: (caseId: number) => Promise<void>
}

export const useDocumentStore = create<DocumentState>((set) => ({
  list: [],
  total: 0,
  byCase: [],
  async fetchList(params = {}) {
    const res: any = await listDocuments(params)
    set({ list: res.data.list, total: res.data.total })
  },
  async fetchByCase(caseId: number) {
    const res: any = await listDocumentsByCase(caseId)
    set({ byCase: res.data })
  },
}))
