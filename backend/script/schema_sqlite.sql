-- SQLite schema for jcourse
-- Generated from entity definitions; run against a fresh SQLite 3 database.
-- Differences from the PostgreSQL version:
--   - SERIAL → INTEGER PRIMARY KEY AUTOINCREMENT
--   - TEXT[] arrays stored as JSON text (matches gorm serializer:json tag)
--   - TIMESTAMPTZ → TEXT (ISO-8601 strings) or INTEGER (unix seconds)
--   - REAL/DOUBLE PRECISION → REAL

PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS departments (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS semesters (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT    NOT NULL,
    can_review INTEGER NOT NULL DEFAULT 0  -- 0 = false, 1 = true
);

CREATE TABLE IF NOT EXISTS teachers (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    code          TEXT    NOT NULL,
    name          TEXT    NOT NULL,
    department    TEXT    NOT NULL,
    title         TEXT    NOT NULL,
    pinyin        TEXT    NOT NULL,
    pinyin_abbr   TEXT    NOT NULL,
    late_semester TEXT    NOT NULL DEFAULT '',
    created_at    TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at    TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS courses (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    code            TEXT    NOT NULL UNIQUE,
    name            TEXT    NOT NULL,
    credit          REAL    NOT NULL,
    department      TEXT    NOT NULL,
    main_teacher_id INTEGER,
    categories      TEXT    NOT NULL DEFAULT '[]',  -- JSON array
    language        TEXT    NOT NULL,
    target_years    TEXT    NOT NULL DEFAULT '[]',  -- JSON array
    review_count    INTEGER NOT NULL DEFAULT 0,
    avg_rating      REAL    NOT NULL DEFAULT 0,
    created_at      TEXT    NOT NULL DEFAULT (datetime('now')),

    FOREIGN KEY (main_teacher_id) REFERENCES teachers(id)
        ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_courses_department   ON courses (department);
CREATE INDEX IF NOT EXISTS idx_courses_main_teacher ON courses (main_teacher_id);

CREATE TABLE IF NOT EXISTS offered_courses (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    course_id  INTEGER NOT NULL,
    semester   TEXT    NOT NULL,
    language   TEXT    NOT NULL,
    target_years TEXT    NOT NULL DEFAULT '[]',       -- JSON array
    categories TEXT    NOT NULL DEFAULT '[]',       -- JSON array

    FOREIGN KEY (course_id) REFERENCES courses(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_offered_courses_course ON offered_courses (course_id);

CREATE TABLE IF NOT EXISTS course_teacher_groups (
    offered_course_id INTEGER NOT NULL,
    teacher_id        INTEGER NOT NULL,

    FOREIGN KEY (offered_course_id) REFERENCES offered_courses(id)
        ON DELETE CASCADE,
    FOREIGN KEY (teacher_id) REFERENCES teachers(id)
        ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_offered_teacher ON course_teacher_groups (offered_course_id, teacher_id);

CREATE TABLE IF NOT EXISTS users (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    username     TEXT    NOT NULL UNIQUE,
    email        TEXT    NOT NULL UNIQUE,
    role         TEXT    NOT NULL,
    password     TEXT    NOT NULL,
    created_at   TEXT    NOT NULL DEFAULT (datetime('now')),
    last_seen_at TEXT    NOT NULL DEFAULT (datetime('now')),
    suspended_at TEXT,
    suspend_till TEXT
);

CREATE TABLE IF NOT EXISTS reviews (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    course_id  INTEGER NOT NULL,
    semester   TEXT    NOT NULL,
    user_id    INTEGER NOT NULL,
    rating     INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    content    TEXT    NOT NULL,
    score     TEXT    NOT NULL,
    created_at TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT    NOT NULL DEFAULT (datetime('now')),

    FOREIGN KEY (course_id) REFERENCES courses(id)
        ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_reviews_course ON reviews (course_id);
CREATE INDEX IF NOT EXISTS idx_reviews_user   ON reviews (user_id);

CREATE TABLE IF NOT EXISTS review_revisions (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    review_id  INTEGER NOT NULL,
    course_id  INTEGER NOT NULL,
    semester   TEXT    NOT NULL,
    user_id    INTEGER NOT NULL,
    rating     INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    content    TEXT    NOT NULL,
    score     TEXT    NOT NULL,
    created_at TEXT    NOT NULL DEFAULT (datetime('now')),

    FOREIGN KEY (review_id) REFERENCES reviews(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_review_revisions_review ON review_revisions (review_id);
CREATE INDEX IF NOT EXISTS idx_review_revisions_course ON review_revisions (course_id);
