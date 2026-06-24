import { Navigate } from 'react-router-dom'
import { useAuthStore } from '@/store/authStore'

interface Props {
  children: React.ReactNode
  roles?: string[]
}

export function ProtectedRoute({ children, roles }: Props) {
  const { token, user } = useAuthStore()

  if (!token) {
    return <Navigate to="/login" replace />
  }

  if (roles && user && !roles.includes(user.role)) {
    return <Navigate to="/admin" replace />
  }

  return <>{children}</>
}