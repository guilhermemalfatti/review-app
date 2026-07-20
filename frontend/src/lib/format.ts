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

export function indicationSummary(aggregates: {
  recommend_count: number
  not_recommend_count: number
  avg_overall: number | null
  last_service_date: string | null
}): string {
  const rec = aggregates.recommend_count
  const not = aggregates.not_recommend_count
  const avg = formatScore(aggregates.avg_overall)
  const last = formatDate(aggregates.last_service_date)
  return `${rec} recomendam · ${not} não · nota média ${avg} · último serviço ${last}`
}
