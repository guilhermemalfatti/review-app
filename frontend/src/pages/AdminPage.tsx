import { useCallback, useEffect, useState } from 'react'
import { api, ApiError } from '../api/client'
import type { AdminUser, PendingProvider, PendingReview } from '../api/types'
import { StatusMessage } from '../components/StatusMessage'

type AdminTab = 'providers' | 'reviews' | 'users'

export function AdminPage() {
  const [tab, setTab] = useState<AdminTab>('providers')
  const [providers, setProviders] = useState<PendingProvider[]>([])
  const [publishedProviders, setPublishedProviders] = useState<PendingProvider[]>([])
  const [reviews, setReviews] = useState<PendingReview[]>([])
  const [users, setUsers] = useState<AdminUser[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)
  const [busyId, setBusyId] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)
  const [tempPasswordNotice, setTempPasswordNotice] = useState<{
    name: string
    password: string
  } | null>(null)

  const loadAll = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const [pendingProviders, approvedProviders, pendingReviews, allUsers] = await Promise.all([
        api.adminPendingProviders(),
        api.adminApprovedProviders(),
        api.adminPendingReviews(),
        api.adminListUsers(),
      ])
      setProviders(pendingProviders)
      setPublishedProviders(approvedProviders)
      setReviews(pendingReviews)
      setUsers(allUsers)
    } catch (err) {
      setError(
        err instanceof ApiError
          ? err.message
          : 'Não foi possível carregar. Tente de novo.',
      )
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadAll()
  }, [loadAll])

  useEffect(() => {
    if (!tempPasswordNotice) return
    setCopied(false)
    const timer = window.setTimeout(() => setTempPasswordNotice(null), 60_000)
    return () => window.clearTimeout(timer)
  }, [tempPasswordNotice])

  function switchTab(next: AdminTab) {
    setTab(next)
    setTempPasswordNotice(null)
    setActionError(null)
    setCopied(false)
  }

  async function runAction(id: string, action: () => Promise<void>, onSuccess: () => void) {
    setBusyId(id)
    setActionError(null)
    try {
      await action()
      onSuccess()
    } catch (err) {
      setActionError(
        err instanceof ApiError ? err.message : 'Não deu certo. Tente de novo.',
      )
    } finally {
      setBusyId(null)
    }
  }

  async function handleResetPassword(user: AdminUser) {
    const confirmed = window.confirm(
      `Criar nova senha para ${user.display_name}?\n\nA senha antiga deixa de funcionar.`,
    )
    if (!confirmed) return

    setBusyId(user.id)
    setActionError(null)
    setTempPasswordNotice(null)
    try {
      const result = await api.adminResetPassword(user.id)
      setTempPasswordNotice({
        name: result.user.display_name,
        password: result.temporary_password,
      })
      setUsers((prev) =>
        prev.map((u) => (u.id === user.id ? { ...u, ...result.user } : u)),
      )
    } catch (err) {
      setActionError(
        err instanceof ApiError ? err.message : 'Não deu certo. Tente de novo.',
      )
    } finally {
      setBusyId(null)
    }
  }

  async function handleCopyPassword() {
    if (!tempPasswordNotice) return
    try {
      await navigator.clipboard.writeText(tempPasswordNotice.password)
      setCopied(true)
    } catch {
      setActionError('Não foi possível copiar a senha.')
    }
  }

  function confirmRemoveProvider(name: string) {
    return window.confirm(
      `Remover “${name}” da lista pública?\n\nO registro fica guardado, mas o prestador deixa de aparecer para os moradores.`,
    )
  }

  return (
    <div className="page page--admin page-enter">
      <header className="form-page-header">
        <h1>Área do administrador</h1>
        <p>Escolha o que quer ver. Depois use Sim ou Não.</p>
      </header>

      <div className="admin-tabs" role="tablist" aria-label="Seções">
        <button
          type="button"
          role="tab"
          aria-selected={tab === 'providers'}
          className={`admin-tab ${tab === 'providers' ? 'admin-tab--active' : ''}`}
          onClick={() => switchTab('providers')}
        >
          Prestadores
          <span className="admin-tab__count">{providers.length + publishedProviders.length}</span>
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={tab === 'reviews'}
          className={`admin-tab ${tab === 'reviews' ? 'admin-tab--active' : ''}`}
          onClick={() => switchTab('reviews')}
        >
          Indicações
          <span className="admin-tab__count">{reviews.length}</span>
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={tab === 'users'}
          className={`admin-tab ${tab === 'users' ? 'admin-tab--active' : ''}`}
          onClick={() => switchTab('users')}
        >
          Moradores
          <span className="admin-tab__count">{users.length}</span>
        </button>
      </div>

      {actionError && <StatusMessage tone="error">{actionError}</StatusMessage>}
      {tempPasswordNotice && tab === 'users' && (
        <StatusMessage tone="success">
          Nova senha de <strong>{tempPasswordNotice.name}</strong>:{' '}
          <code className="temp-password">{tempPasswordNotice.password}</code>
          {' '}
          <button type="button" className="btn btn--ghost" onClick={() => void handleCopyPassword()}>
            {copied ? 'Copiado' : 'Copiar'}
          </button>
          <br />
          Anote e passe para a pessoa. Ela deve trocar ao entrar.
        </StatusMessage>
      )}

      {loading && (
        <div className="state-block" role="status">
          Carregando…
        </div>
      )}

      {error && <StatusMessage tone="error">{error}</StatusMessage>}

      {!loading && !error && tab === 'providers' && (
        <>
          <section className="admin-panel" aria-label="Novos prestadores">
            <h2 className="admin-panel__heading">Aguardando aprovação</h2>
            <p className="admin-panel__help">
              Estes nomes ainda não aparecem na lista. Aprove só se conhecer ou confiar.
            </p>
            {providers.length === 0 ? (
              <p className="admin-empty">Nada para aprovar agora.</p>
            ) : (
              <ul className="admin-list">
                {providers.map((item) => (
                  <li key={item.id} className="admin-card">
                    <h3 className="admin-card__title">{item.name}</h3>
                    <p className="admin-card__meta">{item.category}</p>
                    {item.phone ? <p className="admin-card__meta">Telefone: {item.phone}</p> : null}
                    {item.notes ? <p className="admin-card__note">{item.notes}</p> : null}
                    <div className="admin-card__actions">
                      <button
                        type="button"
                        className="btn btn--primary btn--block"
                        disabled={busyId === item.id}
                        onClick={() =>
                          void runAction(
                            item.id,
                            () => api.adminApproveProvider(item.id),
                            () => {
                              setProviders((prev) => prev.filter((p) => p.id !== item.id))
                              setPublishedProviders((prev) => [...prev, { ...item, status: 'approved' }])
                            },
                          )
                        }
                      >
                        Sim, aprovar
                      </button>
                      <button
                        type="button"
                        className="btn btn--ghost btn--block"
                        disabled={busyId === item.id}
                        onClick={() =>
                          void runAction(
                            item.id,
                            () => api.adminRejectProvider(item.id),
                            () => setProviders((prev) => prev.filter((p) => p.id !== item.id)),
                          )
                        }
                      >
                        Não, rejeitar
                      </button>
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </section>

          <section className="admin-panel" aria-label="Prestadores publicados">
            <h2 className="admin-panel__heading">Na lista pública</h2>
            <p className="admin-panel__help">
              Remover tira o prestador da lista dos moradores. O registro fica guardado.
            </p>
            {publishedProviders.length === 0 ? (
              <p className="admin-empty">Nenhum prestador publicado.</p>
            ) : (
              <ul className="admin-list">
                {publishedProviders.map((item) => (
                  <li key={item.id} className="admin-card">
                    <h3 className="admin-card__title">{item.name}</h3>
                    <p className="admin-card__meta">{item.category}</p>
                    {item.phone ? <p className="admin-card__meta">Telefone: {item.phone}</p> : null}
                    <div className="admin-card__actions">
                      <button
                        type="button"
                        className="btn btn--ghost btn--block"
                        disabled={busyId === item.id}
                        onClick={() => {
                          if (!confirmRemoveProvider(item.name)) return
                          void runAction(
                            item.id,
                            () => api.adminRemoveProvider(item.id),
                            () =>
                              setPublishedProviders((prev) => prev.filter((p) => p.id !== item.id)),
                          )
                        }}
                      >
                        Remover da lista
                      </button>
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </section>
        </>
      )}

      {!loading && !error && tab === 'reviews' && (
        <section className="admin-panel" aria-label="Novas indicações">
          <h2 className="admin-panel__heading">Aguardando aprovação</h2>
          <p className="admin-panel__help">
            Leia o comentário. Se estiver ok, aprove. Se tiver palavras ruins, rejeite.
          </p>
          {reviews.length === 0 ? (
            <p className="admin-empty">Nada para aprovar agora.</p>
          ) : (
            <ul className="admin-list">
              {reviews.map((item) => {
                const who =
                  item.author_display_name ||
                  item.author_label ||
                  item.author_email ||
                  'Morador'
                return (
                  <li key={item.id} className="admin-card">
                    <h3 className="admin-card__title">
                      {item.provider_name ?? 'Prestador'}
                    </h3>
                    <p className="admin-card__meta">
                      {item.recommend ? 'Recomenda' : 'Não recomenda'} · escrito por {who}
                    </p>
                    {item.comment ? (
                      <p className="admin-card__note">“{item.comment}”</p>
                    ) : (
                      <p className="admin-card__meta">Sem comentário</p>
                    )}
                    <div className="admin-card__actions">
                      <button
                        type="button"
                        className="btn btn--primary btn--block"
                        disabled={busyId === item.id}
                        onClick={() =>
                          void runAction(
                            item.id,
                            () => api.adminApproveReview(item.id),
                            () => setReviews((prev) => prev.filter((r) => r.id !== item.id)),
                          )
                        }
                      >
                        Sim, aprovar
                      </button>
                      <button
                        type="button"
                        className="btn btn--ghost btn--block"
                        disabled={busyId === item.id}
                        onClick={() =>
                          void runAction(
                            item.id,
                            () => api.adminRejectReview(item.id),
                            () => setReviews((prev) => prev.filter((r) => r.id !== item.id)),
                          )
                        }
                      >
                        Não, rejeitar
                      </button>
                    </div>
                  </li>
                )
              })}
            </ul>
          )}
        </section>
      )}

      {!loading && !error && tab === 'users' && (
        <section className="admin-panel" aria-label="Moradores">
          <p className="admin-panel__help">
            Se alguém esqueceu a senha, crie uma nova e diga para a pessoa.
          </p>
          {users.length === 0 ? (
            <p className="admin-empty">Nenhum morador cadastrado.</p>
          ) : (
            <ul className="admin-list">
              {users.map((user) => (
                <li key={user.id} className="admin-card">
                  <h2 className="admin-card__title">{user.display_name}</h2>
                  <p className="admin-card__meta">{user.email}</p>
                  <p className="admin-card__meta">
                    {user.role === 'admin' ? 'Administrador' : 'Morador'}
                    {user.must_change_password ? ' · precisa trocar a senha' : ''}
                  </p>
                  <div className="admin-card__actions">
                    <button
                      type="button"
                      className="btn btn--primary btn--block"
                      disabled={busyId === user.id}
                      onClick={() => void handleResetPassword(user)}
                    >
                      Criar nova senha
                    </button>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </section>
      )}
    </div>
  )
}
