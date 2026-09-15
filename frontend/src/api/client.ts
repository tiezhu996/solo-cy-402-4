import request from '@/utils/request'
import type { Client } from '@/types'

export function listClients(params: { page?: number; page_size?: number; keyword?: string }) {
  return request.get('/clients', { params })
}

export function getClient(id: number) {
  return request.get(`/clients/${id}`)
}

export function createClient(data: Partial<Client>) {
  return request.post('/clients', data)
}

export function updateClient(id: number, data: Partial<Client>) {
  return request.put(`/clients/${id}`, data)
}

export function deleteClient(id: number) {
  return request.delete(`/clients/${id}`)
}
