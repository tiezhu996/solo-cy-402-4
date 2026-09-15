import { useAuthStore } from '@/stores/authStore'

export function usePermission() {
  const role = useAuthStore((s) => s.user?.role || '')
  const hasRole = (...roles: string[]) => roles.includes(role)
  return { role, hasRole }
}
