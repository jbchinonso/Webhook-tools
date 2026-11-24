CREATE TABLE categories (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL,
    slug       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX categories_slug_uniq ON categories (slug);

INSERT INTO categories (name, slug) VALUES
    ('Music',        'music'),
    ('Tech',         'tech'),
    ('Sports',       'sports'),
    ('Arts',         'arts'),
    ('Comedy',       'comedy'),
    ('Business',     'business');