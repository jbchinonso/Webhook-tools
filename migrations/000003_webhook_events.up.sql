CREATE TABLE webhook_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    endpoint_id UUID NOT NULL REFERENCES webhook_endpoints(id) ON DELETE CASCADE,
    method      TEXT NOT NULL,
    headers     JSONB NOT NULL DEFAULT '{}',
    body        TEXT NOT NULL DEFAULT '',
    source_ip   TEXT NOT NULL DEFAULT '',
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX webhook_events_endpoint_received_idx
    ON webhook_events (endpoint_id, received_at DESC);
