-- +goose Up
-- Soft-remove providers (admin): keep rows for investigation; hide from public.

ALTER TABLE providers DROP CONSTRAINT IF EXISTS providers_status_check;
ALTER TABLE providers ADD CONSTRAINT providers_status_check
    CHECK (status IN ('pending', 'approved', 'rejected', 'removed'));

-- Allow reusing a provider name after soft-remove.
DROP INDEX IF EXISTS providers_condo_id_name_uidx;
CREATE UNIQUE INDEX providers_condo_id_name_active_uidx
    ON providers (condo_id, name) WHERE status <> 'removed';

-- +goose Down
DROP INDEX IF EXISTS providers_condo_id_name_active_uidx;
CREATE UNIQUE INDEX IF NOT EXISTS providers_condo_id_name_uidx ON providers (condo_id, name);

ALTER TABLE providers DROP CONSTRAINT IF EXISTS providers_status_check;
ALTER TABLE providers ADD CONSTRAINT providers_status_check
    CHECK (status IN ('pending', 'approved', 'rejected'));
