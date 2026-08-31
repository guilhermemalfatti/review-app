export const COMMUNITY_NAME = "Cantegril";

// PIX key shown on /apoiar. Replace with your real key when ready.
export const PIX_KEY = "85476e60-b1dc-4abe-94af-49ccea3abba6";
export const PIX_KEY_LABEL = "Chave PIX";

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
  "Outros",
] as const;

export type Category = (typeof CATEGORIES)[number];
