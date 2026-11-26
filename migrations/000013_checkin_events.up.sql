CREATE TABLE check_in_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id   UUID REFERENCES tickets(id) ON DELETE CASCADE,
    ticket_code TEXT NOT NULL,
    event_id    UUID REFERENCES events(id) ON DELETE CASCADE,
    actor_id    UUID REFERENCES users(id) ON DELETE SET NULL,
    result      TEXT NOT NULL
                CHECK (result IN ('success', 'invalid_code', 'already_used', 'wrong_event', 'outside_window')),
    reason      TEXT NOT NULL DEFAULT '',
    ip_address  TEXT NOT NULL DEFAULT '',
    at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX check_in_events_ticket_idx ON check_in_events (ticket_code);
CREATE INDEX check_in_events_at_idx     ON check_in_events (at DESC);