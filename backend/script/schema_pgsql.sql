-- PostgreSQL 18 schema for jcourse
-- Generated from entity definitions; run against a fresh postgres:16 database.

BEGIN;

CREATE TABLE IF NOT EXISTS departments
(
    id         SERIAL PRIMARY KEY,
    name       TEXT        NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS categories
(
    id         SERIAL PRIMARY KEY,
    name       TEXT        NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS semesters
(
    id         SERIAL PRIMARY KEY,
    name       TEXT        NOT NULL,
    can_review BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS teachers
(
    id            SERIAL PRIMARY KEY,
    code          TEXT        NOT NULL UNIQUE,
    name          TEXT        NOT NULL,
    department    TEXT        NOT NULL,
    title         TEXT        NOT NULL,
    pinyin        TEXT        NOT NULL,
    pinyin_abbr   TEXT        NOT NULL,
    search_vector TSVECTOR    NOT NULL DEFAULT '',
    last_semester TEXT        NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS courses
(
    id              SERIAL PRIMARY KEY,
    code            TEXT             NOT NULL,
    name            TEXT             NOT NULL,
    credit          REAL             NOT NULL,
    department      TEXT             NOT NULL,
    main_teacher_id INTEGER,
    categories      TEXT[],
    language        TEXT             NOT NULL,
    target_years    TEXT[],
    teacher_ids     INTEGER[],
    search_vector   TSVECTOR         NOT NULL DEFAULT '',
    last_semester   TEXT             NOT NULL DEFAULT '',
    rating_count    INTEGER          NOT NULL DEFAULT 0,
    rating_avg      DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ      NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_courses_main_teacher
        FOREIGN KEY (main_teacher_id) REFERENCES teachers (id)
            ON DELETE SET NULL
);

CREATE INDEX idx_courses_department ON courses (department);
CREATE INDEX idx_courses_main_teacher ON courses (main_teacher_id);
CREATE INDEX idx_courses_teacher_ids ON courses USING GIN (teacher_ids);
CREATE INDEX idx_courses_search_vector ON courses USING GIN (search_vector);
CREATE UNIQUE INDEX uniq_courses_code_teacher ON courses (code, main_teacher_id);

CREATE INDEX idx_teachers_search_vector ON teachers USING GIN (search_vector);

CREATE TABLE IF NOT EXISTS offered_courses
(
    id           SERIAL PRIMARY KEY,
    course_id    INTEGER     NOT NULL,
    semester     TEXT        NOT NULL,
    language     TEXT        NOT NULL,
    target_years TEXT[],
    categories   TEXT[],
    teacher_ids  INTEGER[],
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_offered_courses_course
        FOREIGN KEY (course_id) REFERENCES courses (id)
            ON DELETE CASCADE
);

CREATE INDEX idx_offered_courses_course ON offered_courses (course_id);
CREATE INDEX idx_offered_courses_teacher_ids ON offered_courses USING GIN (teacher_ids);

CREATE TABLE IF NOT EXISTS users
(
    id           SERIAL PRIMARY KEY,
    username     TEXT        NOT NULL UNIQUE,
    email        TEXT        NOT NULL UNIQUE,
    role         TEXT        NOT NULL,
    password     TEXT        NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    suspended_at TIMESTAMPTZ,
    suspend_till TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS api_keys
(
    id           SERIAL PRIMARY KEY,
    name         TEXT        NOT NULL,
    key          TEXT        NOT NULL UNIQUE,
    role         TEXT        NOT NULL CHECK (role IN ('system', 'user')),
    user_id      INTEGER     NOT NULL DEFAULT 0,
    last_used_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_api_keys_role_user
        CHECK ((role = 'system' AND user_id = 0) OR (role = 'user' AND user_id > 0))
);

CREATE INDEX idx_api_keys_user ON api_keys (user_id, created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS user_point_records
(
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER     NOT NULL,
    reason      TEXT        NOT NULL,
    amount      INTEGER     NOT NULL,
    description TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_user_point_records_user
        FOREIGN KEY (user_id) REFERENCES users (id)
            ON DELETE CASCADE
);

CREATE INDEX idx_user_point_records_user_created
    ON user_point_records (user_id, created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS point_transfers
(
    id                SERIAL PRIMARY KEY,
    sender_user_id    INTEGER     NOT NULL,
    recipient_user_id INTEGER     NOT NULL,
    amount            INTEGER     NOT NULL,
    fee               INTEGER     NOT NULL,
    fee_payer         TEXT        NOT NULL,
    sender_delta      INTEGER     NOT NULL,
    recipient_delta   INTEGER     NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_point_transfers_sender
        FOREIGN KEY (sender_user_id) REFERENCES users (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_point_transfers_recipient
        FOREIGN KEY (recipient_user_id) REFERENCES users (id)
            ON DELETE CASCADE
);

CREATE INDEX idx_point_transfers_sender_created
    ON point_transfers (sender_user_id, created_at DESC, id DESC);
CREATE INDEX idx_point_transfers_recipient_created
    ON point_transfers (recipient_user_id, created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS reviews
(
    id            SERIAL PRIMARY KEY,
    course_id     INTEGER     NOT NULL,
    semester      TEXT        NOT NULL,
    user_id       INTEGER     NOT NULL,
    rating        INTEGER     NOT NULL CHECK (rating BETWEEN 1 AND 5),
    content       TEXT        NOT NULL,
    score         TEXT        NOT NULL,
    moderator_remark TEXT     NOT NULL DEFAULT '',
    like_count    INTEGER     NOT NULL DEFAULT 0,
    dislike_count INTEGER     NOT NULL DEFAULT 0,
    search_vector TSVECTOR    NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_reviews_course
        FOREIGN KEY (course_id) REFERENCES courses (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_reviews_user
        FOREIGN KEY (user_id) REFERENCES users (id)
            ON DELETE CASCADE
);

CREATE INDEX idx_reviews_course ON reviews (course_id);
CREATE INDEX idx_reviews_user ON reviews (user_id);
CREATE INDEX idx_reviews_search_vector ON reviews USING GIN (search_vector);

CREATE TABLE IF NOT EXISTS review_revisions
(
    id         SERIAL PRIMARY KEY,
    review_id  INTEGER     NOT NULL,
    course_id  INTEGER     NOT NULL,
    semester   TEXT        NOT NULL,
    user_id    INTEGER     NOT NULL,
    rating     INTEGER     NOT NULL CHECK (rating BETWEEN 1 AND 5),
    content    TEXT        NOT NULL,
    score      TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_review_revisions_review
        FOREIGN KEY (review_id) REFERENCES reviews (id)
            ON DELETE CASCADE
);

CREATE INDEX idx_review_revisions_review ON review_revisions (review_id);
CREATE INDEX idx_review_revisions_course ON review_revisions (course_id);

CREATE TABLE IF NOT EXISTS review_votes
(
    review_id  INTEGER     NOT NULL,
    user_id    INTEGER     NOT NULL,
    vote_type  SMALLINT    NOT NULL CHECK (vote_type IN (1, -1)),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (review_id, user_id),

    CONSTRAINT fk_review_votes_review
        FOREIGN KEY (review_id) REFERENCES reviews (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_review_votes_user
        FOREIGN KEY (user_id) REFERENCES users (id)
            ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS course_notifications
(
    user_id    INTEGER     NOT NULL,
    course_id  INTEGER     NOT NULL,
    level      SMALLINT    NOT NULL DEFAULT 0 CHECK (level IN (0, 1, 2)),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (user_id, course_id),

    CONSTRAINT fk_course_notifications_course
        FOREIGN KEY (course_id) REFERENCES courses (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_course_notifications_user
        FOREIGN KEY (user_id) REFERENCES users (id)
            ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS site_daily_stats
(
    stat_date    DATE PRIMARY KEY,
    metrics      JSONB       NOT NULL DEFAULT '{}'::jsonb,
    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS announcements
(
    id         SERIAL PRIMARY KEY,
    title      TEXT        NOT NULL,
    body       TEXT        NOT NULL,
    priority   INTEGER     NOT NULL DEFAULT 0,
    show_start TIMESTAMPTZ NOT NULL,
    show_end   TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_announcements_active ON announcements (priority DESC, created_at DESC)
    WHERE show_start <= NOW() AND show_end >= NOW();

COMMIT;
