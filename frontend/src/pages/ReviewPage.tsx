import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api, ApiError } from '../api/client'
import type { MyReview, ProviderDetail } from '../api/types'
import { StatusMessage } from '../components/StatusMessage'

function scoreToInput(value: number | null | undefined): string {
  return value == null ? '' : String(value)
}

export function ReviewPage() {
  const { id } = useParams<{ id: string }>()
  const [provider, setProvider] = useState<ProviderDetail | null>(null)
  const [existingReview, setExistingReview] = useState<MyReview | null>(null)
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
  const [replaced, setReplaced] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (!id) return
    let cancelled = false

    async function load() {
      setLoadingProvider(true)
      try {
        const data = await api.getProvider(id!)
        if (cancelled) return
        setProvider(data)

        try {
          const mine = await api.getMyReview(id!)
          if (cancelled) return
          setExistingReview(mine)
          setRecommend(mine.recommend)
          setScorePrice(scoreToInput(mine.score_price))
          setScoreQuality(scoreToInput(mine.score_quality))
          setScoreDeadline(scoreToInput(mine.score_deadline))
          setComment(mine.comment ?? '')
          setServiceDate(mine.service_date ?? '')
          setShowName(!mine.is_anonymous)
        } catch (err) {
          if (err instanceof ApiError && err.status === 404) {
            if (!cancelled) setExistingReview(null)
          } else if (!cancelled) {
            setExistingReview(null)
          }
        }
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

    if (existingReview) {
      const ok = window.confirm(
        'Você já enviou uma indicação para este prestador. Enviar de novo substitui a anterior e ela volta para aprovação do administrador. Deseja continuar?',
      )
      if (!ok) return
    }

    setError(null)
    setSuccess(false)
    setReplaced(false)
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
      setReplaced(Boolean(existingReview))
      const mine = await api.getMyReview(id)
      setExistingReview(mine)
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

  const statusLabel =
    existingReview?.status === 'approved'
      ? 'aprovada'
      : existingReview?.status === 'pending'
        ? 'aguardando aprovação'
        : existingReview?.status === 'rejected'
          ? 'rejeitada'
          : null

  return (
    <div className="page page--form page-enter">
      <header className="form-page-header">
        <p className="form-page-header__eyebrow">{provider.category}</p>
        <h1>Indicar {provider.name}</h1>
        <p>Conte como foi o serviço. Sua indicação será publicada após aprovação.</p>
      </header>

      {existingReview && (
        <StatusMessage tone="info">
          Você já tem uma indicação para este prestador
          {statusLabel ? ` (${statusLabel})` : ''}. Enviar de novo{' '}
          <strong>substitui a anterior</strong> e ela volta para a fila de
          aprovação.
        </StatusMessage>
      )}

      {success && (
        <StatusMessage tone="success">
          {replaced
            ? 'Indicação atualizada. Ela aguarda nova aprovação do administrador antes de aparecer publicamente.'
            : 'Indicação enviada. Ela aguarda aprovação do administrador antes de aparecer publicamente.'}
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
            <span>Preço — vale o que cobra?</span>
            <select value={scorePrice} onChange={(e) => setScorePrice(e.target.value)}>
              <option value="">Não informar</option>
              <option value="1">1 — Ruim</option>
              <option value="2">2 — Fraco</option>
              <option value="3">3 — Regular</option>
              <option value="4">4 — Bom</option>
              <option value="5">5 — Excelente</option>
            </select>
          </label>
          <label className="field">
            <span>Qualidade — o serviço ficou bom?</span>
            <select value={scoreQuality} onChange={(e) => setScoreQuality(e.target.value)}>
              <option value="">Não informar</option>
              <option value="1">1 — Ruim</option>
              <option value="2">2 — Fraco</option>
              <option value="3">3 — Regular</option>
              <option value="4">4 — Bom</option>
              <option value="5">5 — Excelente</option>
            </select>
          </label>
          <label className="field">
            <span>Prazo — cumpriu o tempo combinado?</span>
            <select value={scoreDeadline} onChange={(e) => setScoreDeadline(e.target.value)}>
              <option value="">Não informar</option>
              <option value="1">1 — Ruim</option>
              <option value="2">2 — Fraco</option>
              <option value="3">3 — Regular</option>
              <option value="4">4 — Bom</option>
              <option value="5">5 — Excelente</option>
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
          {submitting
            ? 'Enviando…'
            : existingReview
              ? 'Atualizar indicação'
              : 'Enviar indicação'}
        </button>
      </form>

      <p className="auth-switch">
        <Link to={`/providers/${provider.id}`}>Voltar ao prestador</Link>
      </p>
    </div>
  )
}
