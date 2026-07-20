import { useState } from 'react'
import { Link } from 'react-router-dom'
import { api, ApiError } from '../api/client'
import { CATEGORIES } from '../api/types'
import { StatusMessage } from '../components/StatusMessage'

export function SuggestProviderPage() {
  const [name, setName] = useState('')
  const [category, setCategory] = useState<string>(CATEGORIES[0])
  const [phone, setPhone] = useState('')
  const [notes, setNotes] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault()
    setError(null)
    setSuccess(false)
    setSubmitting(true)
    try {
      await api.createProvider({
        name: name.trim(),
        category,
        phone: phone.trim(),
        notes: notes.trim(),
      })
      setSuccess(true)
      setName('')
      setPhone('')
      setNotes('')
      setCategory(CATEGORIES[0])
    } catch (err) {
      setError(
        err instanceof ApiError
          ? err.message
          : 'Não foi possível enviar a sugestão.',
      )
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="page page--form page-enter">
      <header className="form-page-header">
        <h1>Sugerir prestador</h1>
        <p>Indique alguém que você já contratou. A sugestão passa por aprovação.</p>
      </header>

      {success && (
        <StatusMessage tone="success">
          Sugestão enviada. Ela aguarda aprovação do administrador antes de
          aparecer na lista.
        </StatusMessage>
      )}

      <form className="form" onSubmit={(e) => void handleSubmit(e)}>
        {error && <StatusMessage tone="error">{error}</StatusMessage>}

        <label className="field">
          <span>Nome</span>
          <input
            type="text"
            required
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
        </label>

        <label className="field">
          <span>Categoria</span>
          <select value={category} onChange={(e) => setCategory(e.target.value)} required>
            {CATEGORIES.map((item) => (
              <option key={item} value={item}>
                {item}
              </option>
            ))}
          </select>
        </label>

        <label className="field">
          <span>Telefone</span>
          <input
            type="tel"
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
            placeholder="(51) 99999-9999"
          />
        </label>

        <label className="field">
          <span>Observações</span>
          <textarea
            rows={3}
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            placeholder="Ex.: atende no condomínio, especialidade…"
          />
        </label>

        <button type="submit" className="btn btn--primary btn--block" disabled={submitting}>
          {submitting ? 'Enviando…' : 'Enviar sugestão'}
        </button>
      </form>

      <p className="auth-switch">
        <Link to="/">Voltar aos prestadores</Link>
      </p>
    </div>
  )
}
