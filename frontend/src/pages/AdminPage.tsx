import { useCallback, useEffect, useState } from 'react'
import { api, ApiError, unwrapList } from '../api/client'
import type { PendingProvider, PendingReview } from '../api/types'
import { StatusMessage } from '../components/StatusMessage'
import { formatDate } from '../lib/format'

export function AdminPage() {
  const [providers, setProviders] = useState<PendingProvider[]>([])
  const [reviews, setReviews] = useState<PendingReview[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)
  const [busyId, setBusyId] = useState<string | null>(null)

  const loadQueues = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const [providersData, reviewsData] = await Promise.all([
        api.adminPendingProviders(),
        api.adminPendingReviews(),
      ])
      setProviders(unwrapList(providersData))
      setReviews(unwrapList(reviewsData))
    } catch (err) {
      setError(
        err instanceof ApiError
          ? err.message
          : 'Não foi possível carregar as filas de moderação.',
      )
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadQueues()
  }, [loadQueues])

  async function runAction(id: string, action: () => Promise<void>) {
    setBusyId(id)
    setActionError(null)
    try {
      await action()
      await loadQueues()
    } catch (err) {
      setActionError(
        err instanceof ApiError ? err.message : 'Não foi possível concluir a ação.',
      )
    } finally {
      setBusyId(null)
    }
  }

  return (
    <div className="page page--admin page-enter">
      <header className="form-page-header">
        <h1>Moderação</h1>
        <p>Aprove ou rejeite sugestões e indicações pendentes.</p>
      </header>

      {actionError && <StatusMessage tone="error">{actionError}</StatusMessage>}

      {loading && (
        <div className="state-block" role="status">
          Carregando filas…
        </div>
      )}

      {error && <StatusMessage tone="error">{error}</StatusMessage>}

      {!loading && !error && (
        <>
          <section className="admin-queue" aria-labelledby="pending-providers">
            <h2 id="pending-providers">Prestadores pendentes</h2>
            {providers.length === 0 ? (
              <p className="empty-line">Nenhuma sugestão aguardando.</p>
            ) : (
              <ul className="admin-list">
                {providers.map((item) => (
                  <li key={item.id} className="admin-item">
                    <div className="admin-item__body">
                      <strong>{item.name}</strong>
                      <span>{item.category}</span>
                      {item.phone && <span>{item.phone}</span>}
                      {item.notes && <p>{item.notes}</p>}
                    </div>
                    <div className="admin-item__actions">
                      <button
                        type="button"
                        className="btn btn--primary btn--small"
                        disabled={busyId === item.id}
                        onClick={() =>
                          void runAction(item.id, () => api.adminApproveProvider(item.id))
                        }
                      >
                        Aprovar
                      </button>
                      <button
                        type="button"
                        className="btn btn--ghost btn--small"
                        disabled={busyId === item.id}
                        onClick={() =>
                          void runAction(item.id, () => api.adminRejectProvider(item.id))
                        }
                      >
                        Rejeitar
                      </button>
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </section>

          <section className="admin-queue" aria-labelledby="pending-reviews">
            <h2 id="pending-reviews">Indicações pendentes</h2>
            {reviews.length === 0 ? (
              <p className="empty-line">Nenhuma indicação aguardando.</p>
            ) : (
              <ul className="admin-list">
                {reviews.map((item) => (
                  <li key={item.id} className="admin-item">
                    <div className="admin-item__body">
                      <strong>
                        {item.provider_name ?? `Prestador ${item.provider_id}`}
                      </strong>
                      <span>
                        {item.recommend ? 'Recomenda' : 'Não recomenda'}
                        {item.author_label ? ` · ${item.author_label}` : ''}
                      </span>
                      {item.comment && <p>{item.comment}</p>}
                      <span className="admin-item__date">
                        Serviço {formatDate(item.service_date)} · Enviado{' '}
                        {formatDate(item.created_at)}
                      </span>
                    </div>
                    <div className="admin-item__actions">
                      <button
                        type="button"
                        className="btn btn--primary btn--small"
                        disabled={busyId === item.id}
                        onClick={() =>
                          void runAction(item.id, () => api.adminApproveReview(item.id))
                        }
                      >
                        Aprovar
                      </button>
                      <button
                        type="button"
                        className="btn btn--ghost btn--small"
                        disabled={busyId === item.id}
                        onClick={() =>
                          void runAction(item.id, () => api.adminRejectReview(item.id))
                        }
                      >
                        Rejeitar
                      </button>
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </section>
        </>
      )}
    </div>
  )
}
