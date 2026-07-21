import { useState } from 'react'
import { Link, Navigate, useLocation, useNavigate } from 'react-router-dom'
import { ApiError } from '../api/client'
import { useAuth } from '../auth/AuthContext'
import { StatusMessage } from '../components/StatusMessage'

function safeRedirectPath(from: unknown): string {
  if (typeof from === 'string' && from.startsWith('/') && !from.startsWith('//')) {
    return from
  }
  return '/'
}

export function LoginPage() {
  const { login, user, loading, mustChangePassword } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const from = safeRedirectPath((location.state as { from?: string } | null)?.from)

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  if (!loading && user) {
    return <Navigate to={mustChangePassword ? '/change-password' : from} replace />
  }

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      const loggedIn = await login({ email: email.trim(), password })
      navigate(loggedIn.must_change_password ? '/change-password' : from, { replace: true })
    } catch (err) {
      setError(
        err instanceof ApiError
          ? err.message
          : 'Não foi possível entrar. Verifique e-mail e senha.',
      )
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="page page--auth page-enter">
      <header className="auth-header">
        <h1>Entrar</h1>
        <p>Acesse com o e-mail do condomínio para indicar e sugerir prestadores.</p>
      </header>

      <form className="form" onSubmit={(e) => void handleSubmit(e)}>
        {error && <StatusMessage tone="error">{error}</StatusMessage>}

        <label className="field">
          <span>E-mail</span>
          <input
            type="email"
            autoComplete="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
        </label>

        <label className="field">
          <span>Senha</span>
          <input
            type="password"
            autoComplete="current-password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </label>

        <button type="submit" className="btn btn--primary btn--block" disabled={submitting}>
          {submitting ? 'Entrando…' : 'Entrar'}
        </button>
      </form>

      <p className="auth-switch">
        Ainda não tem conta? <Link to="/signup">Criar conta</Link>
      </p>
    </div>
  )
}
