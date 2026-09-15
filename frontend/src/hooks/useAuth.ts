import { useAuthStore } from '@/stores/authStore'

export function useAuth() {
  const auth = useAuthStore()
  const isLoggedIn = !!auth.token
  const role = auth.user?.role || ''
  const isAdmin = role === 'admin'
  const isLawyer = role === 'lawyer' || role === 'admin'
  return { auth, isLoggedIn, role, isAdmin, isLawyer }
}
