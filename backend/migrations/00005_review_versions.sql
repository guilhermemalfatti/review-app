-- +goose Up
-- Append-only reviews: never overwrite rows. Old published versions become superseded.

ALTER TABLE reviews DROP CONSTRAINT IF EXISTS reviews_user_id_provider_id_key;

ALTER TABLE reviews DROP CONSTRAINT IF EXISTS reviews_status_check;
ALTER TABLE reviews ADD CONSTRAINT reviews_status_check
    CHECK (status IN ('pending', 'approved', 'rejected', 'superseded'));

-- At most one live pending submission and one published review per member/provider.
CREATE UNIQUE INDEX reviews_one_pending_per_user_provider
    ON reviews (user_id, provider_id) WHERE status = 'pending';

CREATE UNIQUE INDEX reviews_one_approved_per_user_provider
    ON reviews (user_id, provider_id) WHERE status = 'approved';

CREATE INDEX reviews_user_provider_idx ON reviews (user_id, provider_id);

-- +goose Down
DROP INDEX IF EXISTS reviews_user_provider_idx;
DROP INDEX IF EXISTS reviews_one_approved_per_user_provider;
DROP INDEX IF EXISTS reviews_one_pending_per_user_provider;

ALTER TABLE reviews DROP CONSTRAINT IF EXISTS reviews_status_check;
ALTER TABLE reviews ADD CONSTRAINT reviews_status_check
    CHECK (status IN ('pending', 'approved', 'rejected'));

-- Fails if duplicate (user_id, provider_id) rows exist after Up.
ALTER TABLE reviews ADD CONSTRAINT reviews_user_id_provider_id_key UNIQUE (user_id, provider_id);
