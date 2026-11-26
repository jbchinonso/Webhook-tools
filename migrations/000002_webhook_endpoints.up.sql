CREATE TABLE webhook_endpoints (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL,
    slug       TEXT NOT NULL,
    secret     TEXT NOT NULL DEFAULT '',
    status     TEXT NOT NULL DEFAULT 'active'
               CHECK (status IN ('active', 'disabled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX webhook_endpoints_slug_uniq ON webhook_endpoints (slug);
