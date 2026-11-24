CREATE TABLE payments (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id        UUID NOT NULL REFERENCES bookings(id) ON DELETE RESTRICT,
    idempotency_key   TEXT NOT NULL,
    provider          TEXT NOT NULL DEFAULT 'mock',
    provider_ref      TEXT,
    amount_kobo       BIGINT NOT NULL CHECK (amount_kobo > 0),
    currency          TEXT NOT NULL DEFAULT 'NGN',
    status            TEXT NOT NULL DEFAULT 'initiated'
                      CHECK (status IN ('initiated', 'successful', 'failed', 'cancelled')),
    failure_reason    TEXT,
    initiated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at      TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- The whole idempotency story for payment init lives here. A client that
-- retries a POST with the same key gets the same payment back.
CREATE UNIQUE INDEX payments_idempotency_key_uniq ON payments (idempotency_key);

-- provider_ref is set when the provider acknowledges the init. A webhook
-- arrives keyed by this. UNIQUE means a provider bug that sends the same
-- reference twice can't create two payments.
CREATE UNIQUE INDEX payments_provider_ref_uniq ON payments (provider_ref)
    WHERE provider_ref IS NOT NULL;

-- At most one active (initiated) payment per booking. Once a payment
-- succeeds or fails, a new one can be created. This prevents a client from
-- fanning out 10 pending payments against the same booking.
CREATE UNIQUE INDEX payments_one_active_per_booking
    ON payments (booking_id)
    WHERE status = 'initiated';

CREATE INDEX payments_booking_idx ON payments (booking_id);
CREATE INDEX payments_status_idx  ON payments (status);