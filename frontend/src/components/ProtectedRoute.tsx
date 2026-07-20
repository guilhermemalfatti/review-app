import { Navigate, useLocation } from 'react-router-dom'
import { useAuth } from '../auth/AuthContext'

export function ProtectedRoute({
  children,
  adminOnly = false,
  allowPasswordChange = false,
}: {
  children: React.ReactNode
  adminOnly?: boolean
  allowPasswordChange?: boolean
}) {
  const { user, loading, isAdmin, mustChangePassword } = useAuth()
  const location = useLocation()

  if (loading) {
    return (
      <div className="state-block" role="status">
        Carregando…
      </div>
    )
  }

  if (!user) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />
  }

  if (mustChangePassword && !allowPasswordChange) {
    return <Navigate to="/change-password" replace />
  }

  if (adminOnly && !isAdmin) {
    return <Navigate to="/" replace />
  }

  return children
}
