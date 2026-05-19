CREATE TABLE IF NOT EXISTS course_teacher_groups (
    offered_course_id INT NOT NULL REFERENCES offered_courses(id),
    teacher_id        INT NOT NULL REFERENCES teachers(id),
    PRIMARY KEY (offered_course_id, teacher_id)
);

CREATE INDEX IF NOT EXISTS idx_ctg_teacher_id ON course_teacher_groups(teacher_id);
