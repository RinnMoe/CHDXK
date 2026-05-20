-- PostgreSQL 18 schema for jcourse
-- Generated from entity definitions; run against a fresh postgres:16 database.

BEGIN;

CREATE TABLE IF NOT EXISTS departments (
    id       SERIAL PRIMARY KEY,
    name     TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS semesters (
    id         SERIAL PRIMARY KEY,
    name       TEXT    NOT NULL,
    can_review BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS teachers (
    id            SERIAL PRIMARY KEY,
    code          TEXT    NOT NULL,
    name          TEXT    NOT NULL,
    department    TEXT    NOT NULL,
    title         TEXT    NOT NULL,
    pinyin        TEXT    NOT NULL,
    pinyin_abbr   TEXT    NOT NULL,
    late_semester TEXT    NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS courses (
    id            SERIAL PRIMARY KEY,
    code          TEXT    NOT NULL UNIQUE,
    name          TEXT    NOT NULL,
    credit        REAL    NOT NULL,
    department    TEXT    NOT NULL,
    main_teacher_id INTEGER,
    categories    TEXT[]  NOT NULL DEFAULT '{}',
    language      TEXT    NOT NULL,
    target_years  TEXT[]  NOT NULL DEFAULT '{}',
    teacher_ids   INTEGER[] NOT NULL DEFAULT '{}',
    last_semester TEXT    NOT NULL DEFAULT '',
    rating_count  INTEGER NOT NULL DEFAULT 0,
    rating_avg    DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_courses_main_teacher
        FOREIGN KEY (main_teacher_id) REFERENCES teachers(id)
        ON DELETE SET NULL
);

CREATE INDEX idx_courses_department    ON courses (department);
CREATE INDEX idx_courses_main_teacher  ON courses (main_teacher_id);
CREATE INDEX idx_courses_teacher_ids   ON courses USING GIN (teacher_ids);

CREATE TABLE IF NOT EXISTS offered_courses (
    id         SERIAL PRIMARY KEY,
    course_id  INTEGER NOT NULL,
    semester   TEXT    NOT NULL,
    language   TEXT    NOT NULL,
    target_years TEXT[]  NOT NULL DEFAULT '{}',
    categories TEXT[]  NOT NULL DEFAULT '{}',
    teacher_ids INTEGER[] NOT NULL DEFAULT '{}',

    CONSTRAINT fk_offered_courses_course
        FOREIGN KEY (course_id) REFERENCES courses(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_offered_courses_course ON offered_courses (course_id);
CREATE INDEX idx_offered_courses_teacher_ids ON offered_courses USING GIN (teacher_ids);

CREATE TABLE IF NOT EXISTS users (
    id           SERIAL PRIMARY KEY,
    username     TEXT    NOT NULL UNIQUE,
    email        TEXT    NOT NULL UNIQUE,
    role         TEXT    NOT NULL,
    password     TEXT    NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    suspended_at TIMESTAMPTZ,
    suspend_till TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS reviews (
    id         SERIAL PRIMARY KEY,
    course_id  INTEGER NOT NULL,
    semester   TEXT    NOT NULL,
    user_id    INTEGER NOT NULL,
    rating     INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    content    TEXT    NOT NULL,
    score     TEXT    NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_reviews_course
        FOREIGN KEY (course_id) REFERENCES courses(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_reviews_user
        FOREIGN KEY (user_id) REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_reviews_course ON reviews (course_id);
CREATE INDEX idx_reviews_user  ON reviews (user_id);

CREATE TABLE IF NOT EXISTS review_revisions (
    id         SERIAL PRIMARY KEY,
    review_id  INTEGER NOT NULL,
    course_id  INTEGER NOT NULL,
    semester   TEXT    NOT NULL,
    user_id    INTEGER NOT NULL,
    rating     INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    content    TEXT    NOT NULL,
    score      TEXT    NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_review_revisions_review
        FOREIGN KEY (review_id) REFERENCES reviews(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_review_revisions_review ON review_revisions (review_id);
CREATE INDEX idx_review_revisions_course ON review_revisions (course_id);

COMMIT;
