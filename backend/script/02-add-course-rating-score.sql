ALTER TABLE courses
    ADD COLUMN IF NOT EXISTS rating_score DOUBLE PRECISION NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_courses_rating_score_id
    ON courses (rating_score DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_courses_main_teacher_rating_score
    ON courses (main_teacher_id, rating_score DESC, id DESC);

WITH global_rating AS (
    SELECT COALESCE(AVG(rating), 0)::double precision AS avg_rating
    FROM reviews
), course_rating AS (
    SELECT
        course_id,
        COUNT(*)::double precision AS rating_count,
        AVG(rating)::double precision AS rating_avg
    FROM reviews
    GROUP BY course_id
), computed AS (
    SELECT
        c.id,
        COALESCE(cr.rating_count, 0) AS rating_count,
        COALESCE(cr.rating_avg, 0) AS rating_avg,
        CASE
            WHEN COALESCE(cr.rating_count, 0) = 0 THEN 0
            ELSE ((cr.rating_avg * cr.rating_count) + (5 * gr.avg_rating)) / (cr.rating_count + 5)
        END AS rating_score
    FROM courses AS c
    CROSS JOIN global_rating AS gr
    LEFT JOIN course_rating AS cr ON cr.course_id = c.id
)
UPDATE courses AS c SET
    rating_count = computed.rating_count::integer,
    rating_avg = computed.rating_avg,
    rating_score = computed.rating_score
FROM computed
WHERE computed.id = c.id;

