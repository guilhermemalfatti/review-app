import { CATEGORIES } from '../config'

interface CategoryChipsProps {
  value: string
  onChange: (category: string) => void
}

export function CategoryChips({ value, onChange }: CategoryChipsProps) {
  return (
    <div className="chip-row" role="group" aria-label="Filtrar por categoria">
      <button
        type="button"
        className={`chip ${value === '' ? 'chip--active' : ''}`}
        onClick={() => onChange('')}
      >
        Todas
      </button>
      {CATEGORIES.map((category) => (
        <button
          key={category}
          type="button"
          className={`chip ${value === category ? 'chip--active' : ''}`}
          onClick={() => onChange(category)}
        >
          {category}
        </button>
      ))}
    </div>
  )
}
