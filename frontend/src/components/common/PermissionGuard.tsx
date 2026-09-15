import type { ReactNode } from 'react'
import { usePermission } from '@/hooks/usePermission'

export default function PermissionGuard({ roles, children }: { roles: string[]; children: ReactNode }) {
  const { hasRole } = usePermission()
  if (!hasRole(...roles)) return null
  return <>{children}</>
}
