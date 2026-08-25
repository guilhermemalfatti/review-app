-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS condominio TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS lote TEXT;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_condominio_check;
ALTER TABLE users ADD CONSTRAINT users_condominio_check
  CHECK (condominio IS NULL OR condominio IN ('Fase 1', 'Fase 2', 'Fase 3', 'Fase 4'));

-- +goose Down
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_condominio_check;
ALTER TABLE users DROP COLUMN IF EXISTS lote;
ALTER TABLE users DROP COLUMN IF EXISTS condominio;
