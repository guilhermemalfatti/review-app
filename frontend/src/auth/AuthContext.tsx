import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import { api } from '../api/client'
import type { ChangePasswordPayload, LoginPayload, SignupPayload, User } from '../api/types'

interface AuthContextValue {
  user: User | null
  loading: boolean
  isAdmin: boolean
  mustChangePassword: boolean
  login: (payload: LoginPayload) => Promise<User>
  signup: (payload: SignupPayload) => Promise<User>
  logout: () => Promise<void>
  changePassword: (payload: ChangePasswordPayload) => Promise<void>
  refresh: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  const refresh = useCallback(async () => {
    try {
      const data = await api.me()
      setUser(data.user)
    } catch {
      setUser(null)
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
    return data.user
  }, [])

  const signup = useCallback(async (payload: SignupPayload) => {
    const data = await api.signup(payload)
    setUser(data.user)
    return data.user
  }, [])

  const logout = useCallback(async () => {
    try {
      await api.logout()
    } finally {
      setUser(null)
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
      isAdmin: user?.role === 'admin',
      mustChangePassword: Boolean(user?.must_change_password),
      login,
      signup,
      logout,
      changePassword,
      refresh,
    }),
    [user, loading, login, signup, logout, changePassword, refresh],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return ctx
}
