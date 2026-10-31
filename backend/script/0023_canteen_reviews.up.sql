-- Canteen review events, aggregate summaries, media and reports.
-- Aggregates are persisted so writes only recompute the affected store/user.

CREATE TABLE IF NOT EXISTS canteen_media
(
    id             SERIAL PRIMARY KEY,
    owner_user_id  INTEGER     NOT NULL,
    purpose        TEXT        NOT NULL DEFAULT 'review'
        CHECK (purpose IN ('review', 'dish', 'store_cover', 'report')),
    storage_driver TEXT        NOT NULL DEFAULT 'local',
    storage_key    TEXT        NOT NULL,
    mime           TEXT        NOT NULL DEFAULT '',
    byte_size      INTEGER     NOT NULL DEFAULT 0,
    width          INTEGER     NOT NULL DEFAULT 0,
    height         INTEGER     NOT NULL DEFAULT 0,
    sha256         TEXT        NOT NULL DEFAULT '',
    status         TEXT        NOT NULL DEFAULT 'pending_upload'
        CHECK (status IN ('pending_upload', 'pending', 'approved', 'rejected', 'archived')),
    reviewed_by    INTEGER,
    reviewed_at    TIMESTAMPTZ,
    reject_reason  TEXT        NOT NULL DEFAULT '',
    expires_at     TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_canteen_media_owner
        FOREIGN KEY (owner_user_id) REFERENCES users (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_canteen_media_reviewed_by
        FOREIGN KEY (reviewed_by) REFERENCES users (id)
            ON DELETE SET NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_canteen_media_storage_key
    ON canteen_media (storage_key);
CREATE INDEX IF NOT EXISTS idx_canteen_media_status_created
    ON canteen_media (status, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_canteen_media_owner_created
    ON canteen_media (owner_user_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_canteen_media_sha256
    ON canteen_media (sha256);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_canteen_stores_cover') THEN
        ALTER TABLE canteen_stores
            ADD CONSTRAINT fk_canteen_stores_cover
                FOREIGN KEY (cover_media_id) REFERENCES canteen_media (id)
                    ON DELETE SET NULL;
    END IF;
END
$$;

CREATE TABLE IF NOT EXISTS canteen_review_events
(
    id                    SERIAL PRIMARY KEY,
    store_id              INTEGER     NOT NULL,
    user_id               INTEGER     NOT NULL,
    visited_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    taste                 SMALLINT    NOT NULL CHECK (taste BETWEEN 1 AND 5),
    value                 SMALLINT    NOT NULL CHECK (value BETWEEN 1 AND 5),
    queue                 SMALLINT    NOT NULL CHECK (queue BETWEEN 1 AND 5),
    hygiene               SMALLINT    NOT NULL CHECK (hygiene BETWEEN 1 AND 5),
    service               SMALLINT    NOT NULL CHECK (service BETWEEN 1 AND 5),
    overall_score         NUMERIC(3, 2) NOT NULL,
    score_formula_version SMALLINT    NOT NULL DEFAULT 1,
    content               TEXT        NOT NULL DEFAULT '',
    tags                  TEXT[]      NOT NULL DEFAULT '{}',
    status                TEXT        NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'hidden', 'deleted')),
    like_count            INTEGER     NOT NULL DEFAULT 0,
    dislike_count         INTEGER     NOT NULL DEFAULT 0,
    client_request_id     UUID,
    hidden_reason         TEXT        NOT NULL DEFAULT '',
    hidden_by             INTEGER,
    hidden_at             TIMESTAMPTZ,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_canteen_review_events_store
        FOREIGN KEY (store_id) REFERENCES canteen_stores (id)
            ON DELETE RESTRICT,
    CONSTRAINT fk_canteen_review_events_user
        FOREIGN KEY (user_id) REFERENCES users (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_canteen_review_events_hidden_by
        FOREIGN KEY (hidden_by) REFERENCES users (id)
            ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_canteen_review_events_store_created
    ON canteen_review_events (store_id, status, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_canteen_review_events_user_created
    ON canteen_review_events (user_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_canteen_review_events_store_user_window
    ON canteen_review_events (store_id, user_id, status, visited_at DESC, id DESC);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_canteen_review_events_user_request
    ON canteen_review_events (user_id, client_request_id)
    WHERE client_request_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS canteen_review_dishes
(
    review_event_id INTEGER PRIMARY KEY,
    dish_id         INTEGER     NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_canteen_review_dishes_event
        FOREIGN KEY (review_event_id) REFERENCES canteen_review_events (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_canteen_review_dishes_dish
        FOREIGN KEY (dish_id) REFERENCES canteen_dishes (id)
            ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_canteen_review_dishes_dish
    ON canteen_review_dishes (dish_id);

CREATE TABLE IF NOT EXISTS canteen_user_rating_summaries
(
    store_id             INTEGER     NOT NULL,
    user_id              INTEGER     NOT NULL,
    weighted_score       NUMERIC(3, 2) NOT NULL,
    window_size          SMALLINT    NOT NULL,
    lifetime_event_count INTEGER     NOT NULL DEFAULT 0,
    last_visited_at      TIMESTAMPTZ,
    computed_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (store_id, user_id),

    CONSTRAINT fk_canteen_user_summaries_store
        FOREIGN KEY (store_id) REFERENCES canteen_stores (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_canteen_user_summaries_user
        FOREIGN KEY (user_id) REFERENCES users (id)
            ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_canteen_user_summaries_store
    ON canteen_user_rating_summaries (store_id);

CREATE TABLE IF NOT EXISTS canteen_review_votes
(
    review_event_id INTEGER     NOT NULL,
    user_id         INTEGER     NOT NULL,
    vote_type       SMALLINT    NOT NULL CHECK (vote_type IN (1, -1)),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (review_event_id, user_id),

    CONSTRAINT fk_canteen_review_votes_event
        FOREIGN KEY (review_event_id) REFERENCES canteen_review_events (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_canteen_review_votes_user
        FOREIGN KEY (user_id) REFERENCES users (id)
            ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_canteen_review_votes_user_updated
    ON canteen_review_votes (user_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_canteen_review_votes_event_type
    ON canteen_review_votes (review_event_id, vote_type);

CREATE TABLE IF NOT EXISTS canteen_review_revisions
(
    id              SERIAL PRIMARY KEY,
    review_event_id INTEGER     NOT NULL,
    created_by      INTEGER     NOT NULL,
    visited_at      TIMESTAMPTZ NOT NULL,
    taste           SMALLINT    NOT NULL,
    value           SMALLINT    NOT NULL,
    queue           SMALLINT    NOT NULL,
    hygiene         SMALLINT    NOT NULL,
    service         SMALLINT    NOT NULL,
    overall_score   NUMERIC(3, 2) NOT NULL,
    content         TEXT        NOT NULL DEFAULT '',
    tags            TEXT[]      NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_canteen_review_revisions_event
        FOREIGN KEY (review_event_id) REFERENCES canteen_review_events (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_canteen_review_revisions_user
        FOREIGN KEY (created_by) REFERENCES users (id)
            ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_canteen_review_revisions_event_created
    ON canteen_review_revisions (review_event_id, created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS canteen_dish_review_events
(
    id              SERIAL PRIMARY KEY,
    dish_id         INTEGER     NOT NULL,
    user_id         INTEGER     NOT NULL,
    review_event_id INTEGER,
    taste           SMALLINT    NOT NULL CHECK (taste BETWEEN 1 AND 5),
    portion         SMALLINT    NOT NULL CHECK (portion BETWEEN 1 AND 5),
    value           SMALLINT    NOT NULL CHECK (value BETWEEN 1 AND 5),
    overall_score   NUMERIC(3, 2) NOT NULL,
    content         TEXT        NOT NULL DEFAULT '',
    status          TEXT        NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'hidden', 'deleted')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_canteen_dish_review_events_dish
        FOREIGN KEY (dish_id) REFERENCES canteen_dishes (id)
            ON DELETE RESTRICT,
    CONSTRAINT fk_canteen_dish_review_events_user
        FOREIGN KEY (user_id) REFERENCES users (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_canteen_dish_review_events_review
        FOREIGN KEY (review_event_id) REFERENCES canteen_review_events (id)
            ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_canteen_dish_review_events_dish_created
    ON canteen_dish_review_events (dish_id, status, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_canteen_dish_review_events_dish_user
    ON canteen_dish_review_events (dish_id, user_id, status, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_canteen_dish_review_events_user_created
    ON canteen_dish_review_events (user_id, created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS canteen_dish_rating_summaries
(
    dish_id              INTEGER     NOT NULL,
    user_id              INTEGER     NOT NULL,
    weighted_score       NUMERIC(3, 2) NOT NULL,
    window_size          SMALLINT    NOT NULL,
    lifetime_event_count INTEGER     NOT NULL DEFAULT 0,
    last_reviewed_at     TIMESTAMPTZ,
    computed_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (dish_id, user_id),

    CONSTRAINT fk_canteen_dish_summaries_dish
        FOREIGN KEY (dish_id) REFERENCES canteen_dishes (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_canteen_dish_summaries_user
        FOREIGN KEY (user_id) REFERENCES users (id)
            ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_canteen_dish_summaries_dish
    ON canteen_dish_rating_summaries (dish_id);

CREATE TABLE IF NOT EXISTS canteen_review_media
(
    review_event_id INTEGER  NOT NULL,
    media_id        INTEGER  NOT NULL,
    sort_order      SMALLINT NOT NULL DEFAULT 0,

    PRIMARY KEY (review_event_id, media_id),

    CONSTRAINT fk_canteen_review_media_event
        FOREIGN KEY (review_event_id) REFERENCES canteen_review_events (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_canteen_review_media_media
        FOREIGN KEY (media_id) REFERENCES canteen_media (id)
            ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_canteen_review_media_media
    ON canteen_review_media (media_id);

CREATE TABLE IF NOT EXISTS canteen_reports
(
    id                SERIAL PRIMARY KEY,
    reporter_user_id  INTEGER     NOT NULL,
    subject_type      TEXT        NOT NULL
        CHECK (subject_type IN ('store', 'review', 'dish_review', 'media')),
    subject_id        INTEGER     NOT NULL,
    reason_code       TEXT        NOT NULL,
    detail            TEXT        NOT NULL DEFAULT '',
    status            TEXT        NOT NULL DEFAULT 'open'
        CHECK (status IN ('open', 'reviewing', 'resolved', 'dismissed')),
    handled_by        INTEGER,
    handled_at        TIMESTAMPTZ,
    resolution_note   TEXT        NOT NULL DEFAULT '',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_canteen_reports_reporter
        FOREIGN KEY (reporter_user_id) REFERENCES users (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_canteen_reports_handler
        FOREIGN KEY (handled_by) REFERENCES users (id)
            ON DELETE SET NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_canteen_reports_open_subject
    ON canteen_reports (reporter_user_id, subject_type, subject_id)
    WHERE status IN ('open', 'reviewing');
CREATE INDEX IF NOT EXISTS idx_canteen_reports_status_created
    ON canteen_reports (status, created_at DESC, id DESC);
