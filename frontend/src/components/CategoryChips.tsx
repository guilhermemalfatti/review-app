import { CATEGORIES } from '../config'

interface CategoryChipsProps {
  value: string
  onChange: (category: string) => void
}

export function CategoryChips({ value, onChange }: CategoryChipsProps) {
  return (
    <>
      <label className="category-select" htmlFor="category-filter">
        <span>Categoria</span>
        <select
          id="category-filter"
          value={value}
          onChange={(e) => onChange(e.target.value)}
        >
          <option value="">Todas</option>
          {CATEGORIES.map((category) => (
            <option key={category} value={category}>
              {category}
            </option>
          ))}
        </select>
      </label>

      <div className="chip-row" role="group" aria-label="Filtrar por categoria">
        <button
          type="button"
          className={`chip ${value === '' ? 'chip--active' : ''}`}
          onClick={() => onChange('')}
          aria-pressed={value === ''}
        >
          Todas
        </button>
        {CATEGORIES.map((category) => (
          <button
            key={category}
            type="button"
            className={`chip ${value === category ? 'chip--active' : ''}`}
            onClick={() => onChange(category)}
            aria-pressed={value === category}
          >
            {category}
          </button>
        ))}
      </div>
    </>
  )
}
