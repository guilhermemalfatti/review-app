import { formatScore } from '../lib/format'

function StarRow({ value, max = 5 }: { value: number | null | undefined; max?: number }) {
  if (value == null || Number.isNaN(value)) {
    return (
      <span className="star-row star-row--empty" aria-hidden="true">
        {'☆'.repeat(max)}
      </span>
    )
  }

  const filled = Math.round(Math.min(max, Math.max(0, value)))
  return (
    <span className="star-row" aria-hidden="true">
      <span className="star-row__filled">{'★'.repeat(filled)}</span>
      <span className="star-row__empty">{'☆'.repeat(max - filled)}</span>
    </span>
  )
}

export function ScoreRow({
  label,
  value,
  emphasize = false,
}: {
  label: string
  value: number | null | undefined
  emphasize?: boolean
}) {
  const hasValue = value != null && !Number.isNaN(value)

  return (
    <li className={`score-row${emphasize ? ' score-row--overall' : ''}`}>
      <span className="score-row__label">{label}</span>
      <StarRow value={value} />
      <span className="score-row__num" aria-label={hasValue ? `${formatScore(value)} de 5` : 'Sem nota'}>
        {hasValue ? formatScore(value) : '—'}
      </span>
    </li>
  )
}

export function ScoreInline({
  label,
  value,
}: {
  label: string
  value: number | null | undefined
}) {
  const hasValue = value != null && !Number.isNaN(value)
  if (!hasValue) return null

  return (
    <li className="score-inline">
      <span className="score-inline__label">{label}</span>
      <StarRow value={value} />
      <span className="score-inline__num">{formatScore(value)} de 5</span>
    </li>
  )
}
