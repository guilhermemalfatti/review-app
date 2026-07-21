import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { api, ApiError } from '../api/client'
import type { ChangePasswordPayload, LoginPayload, SignupPayload, User } from '../api/types'

interface AuthContextValue {
  user: User | null
  loading: boolean
  authError: string | null
  isAdmin: boolean
  mustChangePassword: boolean
  login: (payload: LoginPayload) => Promise<User>
  signup: (payload: SignupPayload) => Promise<User>
  logout: () => Promise<void>
  changePassword: (payload: ChangePasswordPayload) => Promise<void>
  refresh: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

const PASSWORD_CHANGE_EXEMPT = new Set(['/change-password', '/login', '/logout'])

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)
  const [authError, setAuthError] = useState<string | null>(null)

  const refresh = useCallback(async () => {
    try {
      const data = await api.me()
      setUser(data.user)
      setAuthError(null)
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        setUser(null)
        setAuthError(null)
      } else {
        setAuthError(
          err instanceof ApiError ? err.message : 'Não foi possível verificar a sessão.',
        )
      }
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void refresh()
  }, [refresh])

  const login = useCallback(async (payload: LoginPayload) => {
    const data = await api.login(payload)
    setUser(data.user)
    setAuthError(null)
    return data.user
  }, [])

  const signup = useCallback(async (payload: SignupPayload) => {
    const data = await api.signup(payload)
    setUser(data.user)
    setAuthError(null)
    return data.user
  }, [])

  const logout = useCallback(async () => {
    try {
      await api.logout()
    } finally {
      setUser(null)
      setAuthError(null)
    }
  }, [])

  const changePassword = useCallback(async (payload: ChangePasswordPayload) => {
    const data = await api.changePassword(payload)
    setUser(data.user)
  }, [])

  const value = useMemo(
    () => ({
      user,
      loading,
      authError,
      isAdmin: user?.role === 'admin',
      mustChangePassword: Boolean(user?.must_change_password),
      login,
      signup,
      logout,
      changePassword,
      refresh,
    }),
    [user, loading, authError, login, signup, logout, changePassword, refresh],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

/** Single gate: redirect when logged-in user must change password. */
export function MustChangePasswordGate({ children }: { children: ReactNode }) {
  const { user, loading, mustChangePassword } = useAuth()
  const location = useLocation()

  if (loading) return children
  if (
    user &&
    mustChangePassword &&
    !PASSWORD_CHANGE_EXEMPT.has(location.pathname)
  ) {
    return <Navigate to="/change-password" replace />
  }
  return children
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return ctx
}
