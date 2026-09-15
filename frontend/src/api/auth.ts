import request from '@/utils/request'

export function login(data: { username: string; password: string }) {
  return request.post('/auth/login', data)
}

export function register(data: Record<string, unknown>) {
  return request.post('/auth/register', data)
}
