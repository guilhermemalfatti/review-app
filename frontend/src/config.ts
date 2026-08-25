export const COMMUNITY_NAME = "Cantegril";

// Frontend UI source of truth for category chips/selects.
// Backend currently accepts free-text categories (no enum enforced).
export const CATEGORIES = [
  "Eletricista",
  "Encanador",
  "Pintor",
  "Pedreiro",
  "Marceneiro",
  "Jardineiro",
  "Limpeza",
  "Ar-condicionado",
  "Telhados",
  "Limpeza de Jardim",
  "Limpeza de Piscina",
  "Limpeza de casas",
  "Limpeza de pátios",
  "Técnico eletrônico",
  "Desentupidora",
  "Funilaria",
  "Marido de aluguel",
  "Manicure",
  "Serralheria",
  "Chaveiro",
  "Junker",
  "Engenheiro",
  "Corretor",
  "Fisioterapia",
  "Pilates",
  "Técnico de informática",
  "Lanches",
  "Impressão",
  "Xerox",
  "Carimbos",
  "Operadora de telefonia",
  "Psicólogo",
  "Outros",
] as const;

export type Category = (typeof CATEGORIES)[number];
