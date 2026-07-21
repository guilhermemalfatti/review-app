-- +goose Up
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;

CREATE UNIQUE INDEX IF NOT EXISTS users_condo_id_email_uidx ON users (condo_id, email);

CREATE UNIQUE INDEX IF NOT EXISTS providers_condo_id_name_uidx ON providers (condo_id, name);

-- +goose Down
DROP INDEX IF EXISTS providers_condo_id_name_uidx;
DROP INDEX IF EXISTS users_condo_id_email_uidx;

CREATE UNIQUE INDEX IF NOT EXISTS users_email_key ON users (email);
