import request from '@/utils/request'
import type { User } from '@/types'

export function getMe() {
  return request.get('/users/me')
}

export function updateProfile(data: Partial<User>) {
  return request.put('/users/me', data)
}

export function listLawyers() {
  return request.get('/users/lawyers')
}

export function listUsers(params: { page?: number; page_size?: number }) {
  return request.get('/users', { params })
}
