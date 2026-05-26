ALTER TABLE courses
    ADD COLUMN IF NOT EXISTS rating_score DOUBLE PRECISION NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_courses_rating_score_id
    ON courses (rating_score DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_courses_main_teacher_rating_score
    ON courses (main_teacher_id, rating_score DESC, id DESC);

WITH global_rating AS (
    SELECT COALESCE(AVG(rating), 0)::double precision AS avg_rating
    FROM reviews
)
UPDATE courses AS c SET
    rating_score = CASE
        WHEN c.rating_count = 0 THEN 0
        ELSE ((c.rating_avg * c.rating_count) + (5 * global_rating.avg_rating)) / (c.rating_count + 5)
    END
FROM global_rating;
