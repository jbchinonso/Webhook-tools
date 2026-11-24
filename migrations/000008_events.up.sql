CREATE TABLE events (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organizer_id  UUID NOT NULL REFERENCES organizers(id) ON DELETE RESTRICT,
    venue_id      UUID REFERENCES venues(id) ON DELETE RESTRICT,
    title         TEXT NOT NULL,
    slug          TEXT NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    status        TEXT NOT NULL DEFAULT 'draft'
                  CHECK (status IN ('draft', 'published', 'cancelled', 'completed')),
    starts_at     TIMESTAMPTZ NOT NULL,
    ends_at       TIMESTAMPTZ NOT NULL,
    timezone      TEXT NOT NULL DEFAULT 'Africa/Lagos',
    cover_image   TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at  TIMESTAMPTZ,
    cancelled_at  TIMESTAMPTZ,

    CONSTRAINT events_time_order CHECK (ends_at > starts_at)
);

CREATE UNIQUE INDEX events_slug_uniq         ON events (slug);
CREATE INDEX        events_organizer_idx     ON events (organizer_id);
CREATE INDEX        events_status_starts_idx ON events (status, starts_at)
    WHERE status = 'published';
CREATE INDEX        events_starts_at_idx     ON events (starts_at);

-- Full-text search over title + description. Uses a GIN index over a
-- generated tsvector. We keep the tsvector materialised so search queries
-- never have to compute it at read time.
ALTER TABLE events
    ADD COLUMN search_tsv tsvector
    GENERATED ALWAYS AS (
        setweight(to_tsvector('english', coalesce(title, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(description, '')), 'B')
    ) STORED;

CREATE INDEX events_search_idx ON events USING GIN (search_tsv);

CREATE TABLE event_categories (
    event_id    UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (event_id, category_id)
);

CREATE INDEX event_categories_category_idx ON event_categories (category_id);