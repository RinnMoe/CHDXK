DROP INDEX IF EXISTS idx_courses_main_teacher_rating_score;
DROP INDEX IF EXISTS idx_courses_rating_score_id;

ALTER TABLE courses
    DROP COLUMN IF EXISTS rating_score;
