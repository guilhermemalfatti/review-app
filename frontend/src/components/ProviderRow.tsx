import { Link } from 'react-router-dom'
import type { ProviderListItem } from '../api/types'
import { indicationSummaryParts } from '../lib/format'

export function ProviderRow({ provider }: { provider: ProviderListItem }) {
  const summary = indicationSummaryParts(provider.aggregates)

  return (
    <Link to={`/providers/${provider.id}`} className="provider-row">
      <div className="provider-row__main">
        <h2 className="provider-row__name">{provider.name}</h2>
        <span className="provider-row__category">{provider.category}</span>
      </div>
      <div className="provider-row__summary">
        <p className="provider-row__summary-line">{summary.recommends}</p>
        <p className="provider-row__summary-line">{summary.score}</p>
        {summary.lastService && (
          <p className="provider-row__summary-line provider-row__summary-line--muted">
            {summary.lastService}
          </p>
        )}
      </div>
    </Link>
  )
}
