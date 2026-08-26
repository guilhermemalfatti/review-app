export function formatDate(value: string | null | undefined): string {
  if (!value) return '—'
  const date = new Date(value.includes('T') ? value : `${value}T12:00:00`)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleDateString('pt-BR', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}

export function formatScore(value: number | null | undefined): string {
  if (value == null || Number.isNaN(value)) return '—'
  return value.toFixed(1).replace('.', ',')
}

/** Plain Portuguese label for a 1–5 score (uses rounded value). */
export function scoreWord(value: number | null | undefined): string | null {
  if (value == null || Number.isNaN(value)) return null
  const n = Math.round(Math.min(5, Math.max(1, value)))
  switch (n) {
    case 1:
      return 'Ruim'
    case 2:
      return 'Fraco'
    case 3:
      return 'Regular'
    case 4:
      return 'Bom'
    case 5:
      return 'Excelente'
    default:
      return null
  }
}

export type IndicationSummaryParts = {
  recommends: string
  score: string
  lastService: string | null
}

export function indicationSummaryParts(aggregates: {
  recommend_count: number
  not_recommend_count: number
  avg_overall: number | null
  last_service_date: string | null
}): IndicationSummaryParts {
  const recommends = `${aggregates.recommend_count} recomendam · ${aggregates.not_recommend_count} não`

  const avg = formatScore(aggregates.avg_overall)
  const word = scoreWord(aggregates.avg_overall)
  const score =
    aggregates.avg_overall == null
      ? 'Sem nota ainda'
      : word
        ? `Nota ${avg} — ${word}`
        : `Nota ${avg} de 5`

  const lastService = aggregates.last_service_date
    ? `Último serviço ${formatDate(aggregates.last_service_date)}`
    : null

  return { recommends, score, lastService }
}
