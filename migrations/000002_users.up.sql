CREATE TABLE users (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email             TEXT NOT NULL,
    password_hash     TEXT NOT NULL,
    role              TEXT NOT NULL CHECK (role IN ('attendee', 'organizer', 'admin')),
    status            TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'deleted')),
    full_name         TEXT NOT NULL DEFAULT '',
    phone             TEXT NOT NULL DEFAULT '',
    email_verified_at TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Case-insensitive uniqueness. Emails are stored lowercased by the service,
-- but the index enforces it even if someone inserts manually.
CREATE UNIQUE INDEX users_email_lower_uniq ON users (LOWER(email));

CREATE INDEX users_role_idx   ON users (role);
CREATE INDEX users_status_idx ON users (status);