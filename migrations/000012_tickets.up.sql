CREATE TABLE tickets (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id     UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    ticket_type_id UUID NOT NULL REFERENCES ticket_types(id) ON DELETE RESTRICT,
    ticket_code    TEXT NOT NULL,
    status         TEXT NOT NULL DEFAULT 'valid'
                   CHECK (status IN ('valid', 'used', 'cancelled')),
    issued_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    checked_in_at  TIMESTAMPTZ,
    checked_in_by  UUID REFERENCES users(id) ON DELETE SET NULL
);

CREATE UNIQUE INDEX tickets_code_uniq    ON tickets (ticket_code);
CREATE INDEX        tickets_booking_idx  ON tickets (booking_id);
CREATE INDEX        tickets_valid_idx    ON tickets (ticket_type_id) WHERE status = 'valid';