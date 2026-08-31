import { useState } from 'react'
import { Navigate, useNavigate } from 'react-router-dom'
import { ApiError } from '../api/client'
import { useAuth } from '../auth/AuthContext'
import { PasswordField } from '../components/PasswordField'
import { StatusMessage } from '../components/StatusMessage'
import { usePageTitle } from '../lib/usePageTitle'

export function ChangePasswordPage() {
  const { user, loading, mustChangePassword, changePassword } = useAuth()
  const navigate = useNavigate()

  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  usePageTitle('Nova senha · Indica')

  if (!loading && !user) {
    return <Navigate to="/login" replace />
  }

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault()
    setError(null)

    if (newPassword.length < 8) {
      setError('A nova senha precisa ter pelo menos 8 caracteres.')
      return
    }
    if (newPassword !== confirmPassword) {
      setError('A confirmação não confere com a nova senha.')
      return
    }

    setSubmitting(true)
    try {
      await changePassword({
        current_password: currentPassword,
        new_password: newPassword,
      })
      navigate('/', { replace: true })
    } catch (err) {
      setError(
        err instanceof ApiError
          ? err.message
          : 'Não foi possível atualizar a senha.',
      )
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="page page--auth page-enter">
      <header className="auth-header">
        <h1>Definir nova senha</h1>
        <p>
          {mustChangePassword
            ? 'Sua senha foi redefinida por um administrador. Escolha uma nova senha para continuar.'
            : 'Altere sua senha de acesso.'}
        </p>
      </header>

      <form className="form" onSubmit={(e) => void handleSubmit(e)}>
        {error && <StatusMessage tone="error">{error}</StatusMessage>}

        <PasswordField
          label="Senha atual (temporária)"
          autoComplete="current-password"
          required
          value={currentPassword}
          onChange={setCurrentPassword}
        />

        <PasswordField
          label="Nova senha"
          autoComplete="new-password"
          required
          minLength={8}
          value={newPassword}
          onChange={setNewPassword}
        />

        <PasswordField
          label="Confirmar nova senha"
          autoComplete="new-password"
          required
          minLength={8}
          value={confirmPassword}
          onChange={setConfirmPassword}
        />

        <button type="submit" className="btn btn--primary btn--block" disabled={submitting}>
          {submitting ? 'Salvando…' : 'Salvar nova senha'}
        </button>
      </form>
    </div>
  )
}
