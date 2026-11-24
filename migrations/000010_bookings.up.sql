CREATE TABLE bookings (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reference     TEXT NOT NULL,
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    event_id      UUID NOT NULL REFERENCES events(id) ON DELETE RESTRICT,
    status        TEXT NOT NULL DEFAULT 'pending'
                  CHECK (status IN ('pending', 'paid', 'cancelled', 'expired')),
    total_kobo    BIGINT NOT NULL CHECK (total_kobo >= 0),
    currency      TEXT NOT NULL DEFAULT 'NGN',
    expires_at    TIMESTAMPTZ NOT NULL,
    paid_at       TIMESTAMPTZ,
    cancelled_at  TIMESTAMPTZ,
    expired_at    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX bookings_reference_uniq ON bookings (reference);
CREATE INDEX bookings_user_idx              ON bookings (user_id, created_at DESC);
CREATE INDEX bookings_event_idx             ON bookings (event_id);
CREATE INDEX bookings_expires_pending_idx   ON bookings (expires_at)
    WHERE status = 'pending';

CREATE TABLE booking_items (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id     UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    ticket_type_id UUID NOT NULL REFERENCES ticket_types(id) ON DELETE RESTRICT,
    quantity       INT NOT NULL CHECK (quantity > 0),
    unit_price_kobo BIGINT NOT NULL CHECK (unit_price_kobo >= 0),
    subtotal_kobo  BIGINT NOT NULL CHECK (subtotal_kobo >= 0),

    CONSTRAINT booking_items_subtotal_consistent
        CHECK (subtotal_kobo = unit_price_kobo * quantity)
);

CREATE INDEX booking_items_booking_idx     ON booking_items (booking_id);
CREATE INDEX booking_items_ticket_type_idx ON booking_items (ticket_type_id);