import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api, ApiError } from '../api/client'
import type { ProviderDetail } from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { ScoreInline, ScoreRow } from '../components/ScoreDisplay'
import { StatusMessage } from '../components/StatusMessage'
import { formatDate, indicationSummaryParts } from '../lib/format'
import { whatsAppURL } from '../lib/phone'
import { usePageTitle } from '../lib/usePageTitle'

export function ProviderDetailPage() {
  const { id } = useParams<{ id: string }>()
  const { user, loading: authLoading } = useAuth()
  const [provider, setProvider] = useState<ProviderDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  usePageTitle(
    provider
      ? `${provider.name} · Indica`
      : error
        ? 'Prestador · Indica'
        : 'Carregando · Indica',
  )

  useEffect(() => {
    if (!id) return
    let cancelled = false

    async function load() {
      setLoading(true)
      setError(null)
      try {
        const data = await api.getProvider(id!)
        if (!cancelled) setProvider(data)
      } catch (err) {
        if (!cancelled) {
          setError(
            err instanceof ApiError
              ? err.message
              : 'Não foi possível carregar este prestador.',
          )
        }
      } finally {
        if (!cancelled) setLoading(false)
      }
    }

    void load()
    return () => {
      cancelled = true
    }
  }, [id])

  if (loading) {
    return (
      <div className="page page-enter">
        <div className="state-block" role="status">
          Carregando prestador…
        </div>
      </div>
    )
  }

  if (error || !provider) {
    return (
      <div className="page page-enter">
        <StatusMessage tone="error">
          {error ?? 'Prestador não encontrado.'}
        </StatusMessage>
        <Link to="/" className="text-link">
          Voltar à lista
        </Link>
      </div>
    )
  }

  const reviews = provider.reviews ?? []
  const aggregates = provider.aggregates
  const summary = indicationSummaryParts(aggregates)
  const hasScores =
    aggregates.avg_price != null ||
    aggregates.avg_quality != null ||
    aggregates.avg_deadline != null ||
    aggregates.avg_overall != null
  const showPhoneGate = !authLoading && !user
  const waURL = provider.phone ? whatsAppURL(provider.phone) : null

  return (
    <div className="page page--detail page-enter">
      <Link to="/" className="text-link text-link--back">
        ← Prestadores
      </Link>

      <header className="detail-header">
        <p className="detail-header__category">{provider.category}</p>
        <h1 className="detail-header__name">{provider.name}</h1>
        <div className="detail-header__summary">
          <p className="detail-header__summary-line">{summary.recommends}</p>
          <p className="detail-header__summary-line">{summary.score}</p>
          {summary.lastService && (
            <p className="detail-header__summary-line">{summary.lastService}</p>
          )}
        </div>
        {provider.phone && (
          <div className="detail-header__phone">
            <p>
              Telefone:{' '}
              <a href={`tel:${provider.phone.replace(/\D/g, '')}`}>{provider.phone}</a>
            </p>
            {waURL && (
              <a
                href={waURL}
                className="btn btn--ghost btn--small"
                target="_blank"
                rel="noopener noreferrer"
              >
                WhatsApp
              </a>
            )}
          </div>
        )}
        {showPhoneGate && (
          <div className="detail-header__phone-gate">
            <p>Entre para ver o telefone</p>
            <Link
              to="/login"
              state={{ from: `/providers/${provider.id}` }}
              className="btn btn--primary"
            >
              Entrar
            </Link>
          </div>
        )}
        {provider.notes && <p className="detail-header__notes">{provider.notes}</p>}
      </header>

      <div className="detail-actions">
        <Link to={`/providers/${provider.id}/review`} className="btn btn--primary">
          Escrever indicação
        </Link>
      </div>

      <section className="score-board" aria-label="Notas dos vizinhos">
        <h2 className="score-board__title">Notas dos vizinhos</h2>
        <p className="score-board__legend">Cada nota vai de 1 (ruim) a 5 (excelente).</p>
        {hasScores ? (
          <ul className="score-board__list">
            <ScoreRow label="Preço" value={aggregates.avg_price} />
            <ScoreRow label="Qualidade" value={aggregates.avg_quality} />
            <ScoreRow label="Prazo" value={aggregates.avg_deadline} />
            <ScoreRow label="Geral" value={aggregates.avg_overall} emphasize />
          </ul>
        ) : (
          <p className="score-board__empty">
            Ainda sem notas — seja o primeiro a indicar.
          </p>
        )}
      </section>

      <section className="reviews-section" aria-labelledby="reviews-heading">
        <h2 id="reviews-heading">O que os vizinhos disseram</h2>

        {reviews.length === 0 ? (
          <div className="state-block">
            <p>Ainda não há indicações públicas para este prestador.</p>
            <p className="state-block__hint">
              Seja o primeiro a contar como foi o serviço.
            </p>
          </div>
        ) : (
          <ol className="review-timeline">
            {reviews.map((review) => (
              <li key={review.id} className="review-item">
                <div className="review-item__meta">
                  <span className="review-item__author">{review.author_label}</span>
                  <span
                    className={`review-item__recommend ${
                      review.recommend
                        ? 'review-item__recommend--yes'
                        : 'review-item__recommend--no'
                    }`}
                  >
                    {review.recommend ? 'Recomenda' : 'Não recomenda'}
                  </span>
                </div>
                <ul className="review-item__scores">
                  <ScoreInline label="Preço" value={review.score_price} />
                  <ScoreInline label="Qualidade" value={review.score_quality} />
                  <ScoreInline label="Prazo" value={review.score_deadline} />
                </ul>
                {review.comment && (
                  <p className="review-item__comment">{review.comment}</p>
                )}
                <p className="review-item__dates">
                  Serviço {formatDate(review.service_date)} · Publicado{' '}
                  {formatDate(review.created_at)}
                </p>
              </li>
            ))}
          </ol>
        )}
      </section>
    </div>
  )
}
