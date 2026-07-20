import { useEffect, useState } from 'react'
import { api, ApiError, unwrapList } from '../api/client'
import type { ProviderListItem } from '../api/types'
import { CategoryChips } from '../components/CategoryChips'
import { ProviderRow } from '../components/ProviderRow'
import { StatusMessage } from '../components/StatusMessage'

export function HomePage() {
  const [providers, setProviders] = useState<ProviderListItem[]>([])
  const [category, setCategory] = useState('')
  const [q, setQ] = useState('')
  const [searchInput, setSearchInput] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    async function load() {
      setLoading(true)
      setError(null)
      try {
        const data = await api.listProviders({
          category: category || undefined,
          q: q || undefined,
        })
        if (!cancelled) {
          setProviders(unwrapList(data))
        }
      } catch (err) {
        if (!cancelled) {
          setError(
            err instanceof ApiError
              ? err.message
              : 'Não foi possível carregar os prestadores.',
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
  }, [category, q])

  function handleSearch(event: React.FormEvent) {
    event.preventDefault()
    setQ(searchInput.trim())
  }

  return (
    <div className="page page--home page-enter">
      <section className="hero">
        <p className="hero__eyebrow">Comunidade Cantegril</p>
        <h1 className="hero__brand">Indica</h1>
        <p className="hero__lead">
          Prestadores de confiança indicados por quem mora ao lado — antes de
          chamar, pergunte aos vizinhos.
        </p>
      </section>

      <section className="list-controls" aria-label="Busca e filtros">
        <form className="search-form" onSubmit={handleSearch}>
          <label className="sr-only" htmlFor="search-q">
            Buscar prestador
          </label>
          <input
            id="search-q"
            type="search"
            placeholder="Buscar por nome…"
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
          />
          <button type="submit" className="btn btn--primary">
            Buscar
          </button>
        </form>
        <CategoryChips value={category} onChange={setCategory} />
      </section>

      {loading && (
        <div className="state-block" role="status">
          Carregando prestadores…
        </div>
      )}

      {error && <StatusMessage tone="error">{error}</StatusMessage>}

      {!loading && !error && providers.length === 0 && (
        <div className="state-block">
          <p>Nenhum prestador encontrado.</p>
          <p className="state-block__hint">
            Tente outra busca ou sugira um prestador que você já contratou.
          </p>
        </div>
      )}

      {!loading && !error && providers.length > 0 && (
        <section className="provider-list" aria-label="Lista de prestadores">
          {providers.map((provider) => (
            <ProviderRow key={provider.id} provider={provider} />
          ))}
        </section>
      )}
    </div>
  )
}
