CREATE TABLE ticket_types (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id           UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    name               TEXT NOT NULL,
    description        TEXT NOT NULL DEFAULT '',
    price_kobo         BIGINT NOT NULL CHECK (price_kobo >= 0),
    currency           TEXT NOT NULL DEFAULT 'NGN',
    total_quantity     INT NOT NULL CHECK (total_quantity > 0),
    available_quantity INT NOT NULL CHECK (available_quantity >= 0),
    max_per_user       INT NOT NULL DEFAULT 10 CHECK (max_per_user > 0),
    sales_start_at     TIMESTAMPTZ,
    sales_end_at       TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT ticket_types_avail_within_total
        CHECK (available_quantity <= total_quantity),
    CONSTRAINT ticket_types_sales_window
        CHECK (sales_end_at IS NULL OR sales_start_at IS NULL OR sales_end_at > sales_start_at)
);

CREATE INDEX ticket_types_event_idx ON ticket_types (event_id);