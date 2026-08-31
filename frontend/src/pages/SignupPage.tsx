import { useEffect, useState } from 'react'
import { Link, Navigate, useNavigate } from 'react-router-dom'
import { api, ApiError } from '../api/client'
import { useAuth } from '../auth/AuthContext'
import { PasswordField } from '../components/PasswordField'
import { StatusMessage } from '../components/StatusMessage'
import { COMMUNITY_NAME } from '../config'
import { usePageTitle } from '../lib/usePageTitle'

export function SignupPage() {
  const { signup, user, loading } = useAuth()
  const navigate = useNavigate()

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [condominio, setCondominio] = useState('')
  const [lote, setLote] = useState('')
  const [inviteCode, setInviteCode] = useState('')
  const [condominioFases, setCondominioFases] = useState<string[]>([])
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  usePageTitle('Criar conta · Indica')

  useEffect(() => {
    let cancelled = false
    void api
      .getConfig()
      .then((config) => {
        if (!cancelled) setCondominioFases(config.condominio_fases)
      })
      .catch(() => {
        if (!cancelled) {
          setError('Não foi possível carregar as opções de condomínio.')
        }
      })
    return () => {
      cancelled = true
    }
  }, [])

  if (!loading && user) {
    return <Navigate to="/" replace />
  }

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      await signup({
        email: email.trim(),
        password,
        display_name: displayName.trim(),
        condominio,
        lote: lote.trim(),
        invite_code: inviteCode.trim(),
      })
      navigate('/', { replace: true })
    } catch (err) {
      setError(
        err instanceof ApiError
          ? err.message
          : 'Não foi possível criar a conta. Confira o código de convite.',
      )
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="page page--auth page-enter">
      <header className="auth-header">
        <h1>Criar conta</h1>
        <p>Use o código de convite do {COMMUNITY_NAME} para entrar na comunidade Indica.</p>
      </header>

      <form className="form" onSubmit={(e) => void handleSubmit(e)}>
        {error && <StatusMessage tone="error">{error}</StatusMessage>}

        <label className="field">
          <span>Nome para exibição</span>
          <input
            type="text"
            autoComplete="name"
            required
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
          />
        </label>

        <label className="field">
          <span>Condomínio</span>
          <select
            required
            value={condominio}
            onChange={(e) => setCondominio(e.target.value)}
            disabled={condominioFases.length === 0}
          >
            <option value="" disabled>
              Selecione a fase
            </option>
            {condominioFases.map((fase) => (
              <option key={fase} value={fase}>
                {fase}
              </option>
            ))}
          </select>
        </label>

        <label className="field">
          <span>Lote</span>
          <input
            type="text"
            required
            maxLength={32}
            value={lote}
            onChange={(e) => setLote(e.target.value)}
          />
        </label>

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

        <PasswordField
          label="Senha"
          autoComplete="new-password"
          required
          minLength={8}
          value={password}
          onChange={setPassword}
        />

        <label className="field">
          <span>Código de convite</span>
          <input
            type="text"
            required
            value={inviteCode}
            onChange={(e) => setInviteCode(e.target.value)}
          />
        </label>

        <button type="submit" className="btn btn--primary btn--block" disabled={submitting}>
          {submitting ? 'Criando…' : 'Criar conta'}
        </button>
      </form>

      <p className="auth-switch">
        Já tem conta? <Link to="/login">Entrar</Link>
      </p>
    </div>
  )
}
