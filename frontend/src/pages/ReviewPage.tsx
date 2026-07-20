import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api, ApiError } from '../api/client'
import type { ProviderDetail } from '../api/types'
import { StatusMessage } from '../components/StatusMessage'

export function ReviewPage() {
  const { id } = useParams<{ id: string }>()
  const [provider, setProvider] = useState<ProviderDetail | null>(null)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [loadingProvider, setLoadingProvider] = useState(true)

  const [recommend, setRecommend] = useState(true)
  const [scorePrice, setScorePrice] = useState('')
  const [scoreQuality, setScoreQuality] = useState('')
  const [scoreDeadline, setScoreDeadline] = useState('')
  const [comment, setComment] = useState('')
  const [serviceDate, setServiceDate] = useState('')
  const [showName, setShowName] = useState(true)

  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (!id) return
    let cancelled = false

    async function load() {
      setLoadingProvider(true)
      try {
        const data = await api.getProvider(id!)
        if (!cancelled) setProvider(data)
      } catch (err) {
        if (!cancelled) {
          setLoadError(
            err instanceof ApiError
              ? err.message
              : 'Não foi possível carregar o prestador.',
          )
        }
      } finally {
        if (!cancelled) setLoadingProvider(false)
      }
    }

    void load()
    return () => {
      cancelled = true
    }
  }, [id])

  function optionalScore(value: string): number | undefined {
    if (!value) return undefined
    const n = Number(value)
    if (Number.isNaN(n) || n < 1 || n > 5) return undefined
    return n
  }

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault()
    if (!id) return
    setError(null)
    setSuccess(false)
    setSubmitting(true)

    try {
      const payload = {
        is_anonymous: !showName,
        recommend,
        comment: comment.trim(),
        ...(optionalScore(scorePrice) != null
          ? { score_price: optionalScore(scorePrice) }
          : {}),
        ...(optionalScore(scoreQuality) != null
          ? { score_quality: optionalScore(scoreQuality) }
          : {}),
        ...(optionalScore(scoreDeadline) != null
          ? { score_deadline: optionalScore(scoreDeadline) }
          : {}),
        ...(serviceDate ? { service_date: serviceDate } : {}),
      }

      await api.createReview(id, payload)
      setSuccess(true)
      setComment('')
      setScorePrice('')
      setScoreQuality('')
      setScoreDeadline('')
      setServiceDate('')
      setRecommend(true)
      setShowName(true)
    } catch (err) {
      setError(
        err instanceof ApiError
          ? err.message
          : 'Não foi possível enviar a indicação.',
      )
    } finally {
      setSubmitting(false)
    }
  }

  if (loadingProvider) {
    return (
      <div className="page page-enter">
        <div className="state-block" role="status">
          Carregando…
        </div>
      </div>
    )
  }

  if (loadError || !provider) {
    return (
      <div className="page page-enter">
        <StatusMessage tone="error">{loadError ?? 'Prestador não encontrado.'}</StatusMessage>
        <Link to="/" className="text-link">
          Voltar à lista
        </Link>
      </div>
    )
  }

  return (
    <div className="page page--form page-enter">
      <header className="form-page-header">
        <p className="form-page-header__eyebrow">{provider.category}</p>
        <h1>Indicar {provider.name}</h1>
        <p>Conte como foi o serviço. Sua indicação será publicada após aprovação.</p>
      </header>

      {success && (
        <StatusMessage tone="success">
          Indicação enviada. Ela aguarda aprovação do administrador antes de
          aparecer publicamente.
        </StatusMessage>
      )}

      <form className="form" onSubmit={(e) => void handleSubmit(e)}>
        {error && <StatusMessage tone="error">{error}</StatusMessage>}

        <fieldset className="field-group">
          <legend>Você recomenda?</legend>
          <div className="toggle-pair" role="group" aria-label="Recomendação">
            <button
              type="button"
              className={`toggle-option ${recommend ? 'toggle-option--active' : ''}`}
              onClick={() => setRecommend(true)}
              aria-pressed={recommend}
            >
              Sim, recomendo
            </button>
            <button
              type="button"
              className={`toggle-option ${!recommend ? 'toggle-option--active toggle-option--warn' : ''}`}
              onClick={() => setRecommend(false)}
              aria-pressed={!recommend}
            >
              Não recomendo
            </button>
          </div>
        </fieldset>

        <div className="score-fields">
          <label className="field">
            <span>Preço (1–5)</span>
            <select value={scorePrice} onChange={(e) => setScorePrice(e.target.value)}>
              <option value="">Opcional</option>
              {[1, 2, 3, 4, 5].map((n) => (
                <option key={n} value={n}>
                  {n}
                </option>
              ))}
            </select>
          </label>
          <label className="field">
            <span>Qualidade (1–5)</span>
            <select value={scoreQuality} onChange={(e) => setScoreQuality(e.target.value)}>
              <option value="">Opcional</option>
              {[1, 2, 3, 4, 5].map((n) => (
                <option key={n} value={n}>
                  {n}
                </option>
              ))}
            </select>
          </label>
          <label className="field">
            <span>Prazo (1–5)</span>
            <select value={scoreDeadline} onChange={(e) => setScoreDeadline(e.target.value)}>
              <option value="">Opcional</option>
              {[1, 2, 3, 4, 5].map((n) => (
                <option key={n} value={n}>
                  {n}
                </option>
              ))}
            </select>
          </label>
        </div>

        <label className="field">
          <span>Comentário</span>
          <textarea
            rows={4}
            value={comment}
            onChange={(e) => setComment(e.target.value)}
            placeholder="Como foi o atendimento, pontualidade, qualidade…"
          />
        </label>

        <label className="field">
          <span>Data do serviço</span>
          <input
            type="date"
            value={serviceDate}
            onChange={(e) => setServiceDate(e.target.value)}
          />
        </label>

        <fieldset className="field-group">
          <legend>Privacidade</legend>
          <div className="privacy-toggle">
            <button
              type="button"
              className={`privacy-switch ${showName ? 'privacy-switch--named' : 'privacy-switch--anon'}`}
              onClick={() => setShowName((v) => !v)}
              aria-pressed={showName}
            >
              <span className="privacy-switch__knob" aria-hidden="true" />
              <span className="privacy-switch__label">
                {showName ? 'Mostrar meu nome' : 'Publicar como anônimo'}
              </span>
            </button>
          </div>
        </fieldset>

        <button type="submit" className="btn btn--primary btn--block" disabled={submitting}>
          {submitting ? 'Enviando…' : 'Enviar indicação'}
        </button>
      </form>

      <p className="auth-switch">
        <Link to={`/providers/${provider.id}`}>Voltar ao prestador</Link>
      </p>
    </div>
  )
}
