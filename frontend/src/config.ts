export const COMMUNITY_NAME = 'Cantegril'

// Frontend UI source of truth for category chips/selects.
// Backend currently accepts free-text categories (no enum enforced).
export const CATEGORIES = [
  'Eletricista',
  'Encanador',
  'Pintor',
  'Pedreiro',
  'Marceneiro',
  'Jardineiro',
  'Limpeza',
  'Ar-condicionado',
  'Outros',
] as const

export type Category = (typeof CATEGORIES)[number]
