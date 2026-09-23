CREATE TABLE IF NOT EXISTS course_sections
(
    id                SERIAL PRIMARY KEY,
    offered_course_id INTEGER     NOT NULL,
    source_lesson_id  TEXT        NOT NULL DEFAULT '',
    lesson_no         TEXT        NOT NULL DEFAULT '',
    course_name       TEXT        NOT NULL,
    course_type       TEXT        NOT NULL DEFAULT '',
    credit            REAL        NOT NULL DEFAULT 0,
    department        TEXT        NOT NULL DEFAULT '',
    language          TEXT        NOT NULL DEFAULT '',
    campus            TEXT        NOT NULL DEFAULT '',
    assessment        TEXT        NOT NULL DEFAULT '',
    teach_class       TEXT        NOT NULL DEFAULT '',
    actual_count      INTEGER,
    limit_count       INTEGER,
    course_period     TEXT        NOT NULL DEFAULT '',
    total_hours       REAL,
    weekly_hours      REAL,
    week_range        TEXT        NOT NULL DEFAULT '',
    start_week        INTEGER,
    end_week          INTEGER,
    week_count        INTEGER,
    remark            TEXT        NOT NULL DEFAULT '',
    syllabus_url      TEXT        NOT NULL DEFAULT '',
    teacher_ids       INTEGER[]   NOT NULL DEFAULT '{}',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_course_sections_offered_course
        FOREIGN KEY (offered_course_id) REFERENCES offered_courses (id)
            ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_course_sections_offered_course
    ON course_sections (offered_course_id, id);
CREATE INDEX IF NOT EXISTS idx_course_sections_course_name
    ON course_sections (course_name);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_course_sections_source_lesson
    ON course_sections (offered_course_id, source_lesson_id)
    WHERE source_lesson_id <> '';
