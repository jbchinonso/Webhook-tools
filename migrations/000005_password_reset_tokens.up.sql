CREATE TABLE password_reset_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash  TEXT NOT NULL,
    attempts   INT NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX prt_user_id_idx    ON password_reset_tokens (user_id);
CREATE INDEX prt_code_hash_idx  ON password_reset_tokens (code_hash);

-- Enforce at most one active (unused, unexpired) reset row per user at a
-- time. Requesting a new code invalidates the previous one.
CREATE UNIQUE INDEX prt_one_active_per_user
    ON password_reset_tokens (user_id)
    WHERE used_at IS NULL;