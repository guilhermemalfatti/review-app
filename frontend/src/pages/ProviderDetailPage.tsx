import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api, ApiError } from '../api/client'
import type { ProviderDetail } from '../api/types'
import { ScoreCard, ScoreInline } from '../components/ScoreDisplay'
import { StatusMessage } from '../components/StatusMessage'
import { formatDate, indicationSummary } from '../lib/format'

export function ProviderDetailPage() {
  const { id } = useParams<{ id: string }>()
  const [provider, setProvider] = useState<ProviderDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

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

  return (
    <div className="page page--detail page-enter">
      <Link to="/" className="text-link text-link--back">
        ← Prestadores
      </Link>

      <header className="detail-header">
        <p className="detail-header__category">{provider.category}</p>
        <h1 className="detail-header__name">{provider.name}</h1>
        <p className="detail-header__summary">
          {indicationSummary(provider.aggregates)}
        </p>
        {provider.phone && (
          <p className="detail-header__phone">
            Telefone:{' '}
            <a href={`tel:${provider.phone.replace(/\D/g, '')}`}>{provider.phone}</a>
          </p>
        )}
        {provider.notes && <p className="detail-header__notes">{provider.notes}</p>}
      </header>

      <section className="score-board" aria-label="Notas dos vizinhos">
        <h2 className="score-board__title">Notas dos vizinhos</h2>
        <p className="score-board__legend">Cada nota vai de 1 (ruim) a 5 (excelente).</p>
        <div className="score-board__grid">
          <ScoreCard
            label="Preço"
            hint="Vale o que cobra?"
            value={provider.aggregates.avg_price}
          />
          <ScoreCard
            label="Qualidade"
            hint="O serviço ficou bom?"
            value={provider.aggregates.avg_quality}
          />
          <ScoreCard
            label="Prazo"
            hint="Cumpriu o tempo combinado?"
            value={provider.aggregates.avg_deadline}
          />
          <ScoreCard
            label="Geral"
            hint="Média de tudo"
            value={provider.aggregates.avg_overall}
          />
        </div>
      </section>

      <div className="detail-actions">
        <Link to={`/providers/${provider.id}/review`} className="btn btn--primary">
          Escrever indicação
        </Link>
      </div>

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
