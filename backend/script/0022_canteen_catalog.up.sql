-- Canteen catalog: campus halls, stores and dishes.
-- Kept separate from courses/reviews so the course domain stays untouched.

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS canteen_halls
(
    id                SERIAL PRIMARY KEY,
    campus_key        TEXT        NOT NULL,
    name              TEXT        NOT NULL,
    short_name        TEXT        NOT NULL DEFAULT '',
    address           TEXT        NOT NULL DEFAULT '',
    latitude          NUMERIC(9, 6),
    longitude         NUMERIC(9, 6),
    external_place_id TEXT,
    status            TEXT        NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'offline', 'hidden')),
    display_order     INTEGER     NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_canteen_halls_campus_name
    ON canteen_halls (campus_key, name);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_canteen_halls_external_place
    ON canteen_halls (external_place_id)
    WHERE external_place_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_canteen_halls_campus_order
    ON canteen_halls (campus_key, display_order, id);

CREATE TABLE IF NOT EXISTS canteen_stores
(
    id                   SERIAL PRIMARY KEY,
    hall_id              INTEGER     NOT NULL,
    name                 TEXT        NOT NULL,
    normalized_name      TEXT        NOT NULL,
    floor                TEXT        NOT NULL DEFAULT '',
    categories           TEXT[]      NOT NULL DEFAULT '{}',
    status               TEXT        NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'active', 'offline', 'rejected', 'merged')),
    merged_into_store_id INTEGER,
    cover_media_id       INTEGER,
    rating_count         INTEGER     NOT NULL DEFAULT 0,
    rating_avg           DOUBLE PRECISION NOT NULL DEFAULT 0,
    rating_score         DOUBLE PRECISION NOT NULL DEFAULT 0,
    taste_avg            DOUBLE PRECISION NOT NULL DEFAULT 0,
    value_avg            DOUBLE PRECISION NOT NULL DEFAULT 0,
    queue_avg            DOUBLE PRECISION NOT NULL DEFAULT 0,
    hygiene_avg          DOUBLE PRECISION NOT NULL DEFAULT 0,
    service_avg          DOUBLE PRECISION NOT NULL DEFAULT 0,
    review_event_count   INTEGER     NOT NULL DEFAULT 0,
    last_review_at       TIMESTAMPTZ,
    rating_updated_at    TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_canteen_stores_hall
        FOREIGN KEY (hall_id) REFERENCES canteen_halls (id)
            ON DELETE RESTRICT,
    CONSTRAINT fk_canteen_stores_merged
        FOREIGN KEY (merged_into_store_id) REFERENCES canteen_stores (id)
            ON DELETE SET NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_canteen_stores_hall_name
    ON canteen_stores (hall_id, normalized_name)
    WHERE status <> 'merged';
CREATE INDEX IF NOT EXISTS idx_canteen_stores_hall_status
    ON canteen_stores (hall_id, status, id);
CREATE INDEX IF NOT EXISTS idx_canteen_stores_name_trgm
    ON canteen_stores USING GIN (normalized_name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_canteen_stores_categories
    ON canteen_stores USING GIN (categories);
CREATE INDEX IF NOT EXISTS idx_canteen_stores_rank
    ON canteen_stores (rating_score DESC, rating_count DESC, id DESC)
    WHERE status = 'active';

CREATE TABLE IF NOT EXISTS canteen_dishes
(
    id                  SERIAL PRIMARY KEY,
    store_id            INTEGER     NOT NULL,
    name                TEXT        NOT NULL,
    normalized_name     TEXT        NOT NULL,
    status              TEXT        NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'active', 'hidden', 'rejected', 'merged')),
    merged_into_dish_id INTEGER,
    rating_count        INTEGER     NOT NULL DEFAULT 0,
    rating_avg          DOUBLE PRECISION NOT NULL DEFAULT 0,
    rating_score        DOUBLE PRECISION NOT NULL DEFAULT 0,
    taste_avg           DOUBLE PRECISION NOT NULL DEFAULT 0,
    portion_avg         DOUBLE PRECISION NOT NULL DEFAULT 0,
    value_avg           DOUBLE PRECISION NOT NULL DEFAULT 0,
    review_event_count  INTEGER     NOT NULL DEFAULT 0,
    rating_updated_at   TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_canteen_dishes_store
        FOREIGN KEY (store_id) REFERENCES canteen_stores (id)
            ON DELETE RESTRICT,
    CONSTRAINT fk_canteen_dishes_merged
        FOREIGN KEY (merged_into_dish_id) REFERENCES canteen_dishes (id)
            ON DELETE SET NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_canteen_dishes_store_name
    ON canteen_dishes (store_id, normalized_name)
    WHERE status <> 'merged';
CREATE INDEX IF NOT EXISTS idx_canteen_dishes_store_status
    ON canteen_dishes (store_id, status, id);
CREATE INDEX IF NOT EXISTS idx_canteen_dishes_name_trgm
    ON canteen_dishes USING GIN (normalized_name gin_trgm_ops);

CREATE TABLE IF NOT EXISTS canteen_dish_aliases
(
    id              SERIAL PRIMARY KEY,
    dish_id         INTEGER     NOT NULL,
    name            TEXT        NOT NULL,
    normalized_name TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_canteen_dish_aliases_dish
        FOREIGN KEY (dish_id) REFERENCES canteen_dishes (id)
            ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_canteen_dish_aliases_dish_name
    ON canteen_dish_aliases (dish_id, normalized_name);
