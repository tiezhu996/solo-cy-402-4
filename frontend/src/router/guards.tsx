import type { ReactNode } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'

export function RequireAuth({ children }: { children: ReactNode }) {
  const token = useAuthStore((s) => s.token)
  const location = useLocation()
  if (!token) {
    return <Navigate to="/login" state={{ from: location }} replace />
  }
  return <>{children}</>
}

export function RequireRole({ roles, children }: { roles: string[]; children: ReactNode }) {
  const role = useAuthStore((s) => s.user?.role || '')
  if (!roles.includes(role)) {
    return <Navigate to="/cases" replace />
  }
  return <>{children}</>
}
