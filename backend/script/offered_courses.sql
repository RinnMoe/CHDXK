CREATE TABLE IF NOT EXISTS offered_courses (
    id        SERIAL PRIMARY KEY,
    course_id INT NOT NULL REFERENCES courses(id),
    semester  VARCHAR(32) NOT NULL,
    language  VARCHAR(32) NOT NULL DEFAULT '',
    grade     VARCHAR(32) NOT NULL DEFAULT '',
    UNIQUE(course_id, semester)
);

CREATE INDEX IF NOT EXISTS idx_oc_course_id ON offered_courses(course_id);
