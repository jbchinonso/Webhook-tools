CREATE TABLE venues (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL,
    address    TEXT NOT NULL,
    city       TEXT NOT NULL,
    country    TEXT NOT NULL DEFAULT 'Nigeria',
    capacity   INT  NOT NULL CHECK (capacity > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX venues_city_idx ON venues (LOWER(city));