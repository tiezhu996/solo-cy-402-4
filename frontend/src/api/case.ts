import request from '@/utils/request'
import type { CaseItem } from '@/types'

export function listCases(params: Record<string, unknown>) {
  return request.get('/cases', { params })
}

export function getCase(id: number) {
  return request.get(`/cases/${id}`)
}

export function createCase(data: Partial<CaseItem>) {
  return request.post('/cases', data)
}

export function updateCase(id: number, data: Partial<CaseItem>) {
  return request.put(`/cases/${id}`, data)
}

export function changeCaseStatus(id: number, status: string) {
  return request.post(`/cases/${id}/status`, { status })
}

export function assignLawyer(id: number, data: { lead_lawyer_id: number; co_lawyer_ids?: number[] }) {
  return request.post(`/cases/${id}/assign`, data)
}
