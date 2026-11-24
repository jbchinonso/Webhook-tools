CREATE TABLE sessions (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id            UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash TEXT NOT NULL,
    user_agent         TEXT NOT NULL DEFAULT '',
    ip_address         TEXT NOT NULL DEFAULT '',
    expires_at         TIMESTAMPTZ NOT NULL,
    revoked_at         TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- The hash is looked up on every refresh. Unique because a collision would
-- mean two sessions accidentally share a token — impossible in practice but
-- worth enforcing.
CREATE UNIQUE INDEX sessions_refresh_token_hash_uniq ON sessions (refresh_token_hash);

CREATE INDEX sessions_user_id_idx    ON sessions (user_id);
CREATE INDEX sessions_expires_at_idx ON sessions (expires_at)
    WHERE revoked_at IS NULL;