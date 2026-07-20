import { formatScore, scoreWord } from '../lib/format'

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

export function ScoreCard({
  label,
  hint,
  value,
}: {
  label: string
  hint: string
  value: number | null | undefined
}) {
  const hasValue = value != null && !Number.isNaN(value)
  const word = scoreWord(value)

  return (
    <div className="score-card">
      <p className="score-card__label">{label}</p>
      <p className="score-card__hint">{hint}</p>
      <StarRow value={value} />
      <p className="score-card__value">
        {hasValue ? (
          <>
            <strong>{formatScore(value)}</strong>
            <span> de 5</span>
            {word ? <span className="score-card__word"> — {word}</span> : null}
          </>
        ) : (
          <span className="score-card__missing">Sem notas ainda</span>
        )}
      </p>
    </div>
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
