import { Link } from 'react-router-dom'
import type { ProviderListItem } from '../api/types'
import { indicationSummary } from '../lib/format'

export function ProviderRow({ provider }: { provider: ProviderListItem }) {
  return (
    <Link to={`/providers/${provider.id}`} className="provider-row">
      <div className="provider-row__main">
        <h2 className="provider-row__name">{provider.name}</h2>
        <span className="provider-row__category">{provider.category}</span>
      </div>
      <p className="provider-row__summary">{indicationSummary(provider.aggregates)}</p>
    </Link>
  )
}
