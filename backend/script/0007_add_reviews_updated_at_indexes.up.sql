DROP INDEX IF EXISTS idx_reviews_user_id_desc;
DROP INDEX IF EXISTS idx_reviews_course_rating_created;
DROP INDEX IF EXISTS idx_reviews_course_semester_created;
DROP INDEX IF EXISTS idx_reviews_course_created;
DROP INDEX IF EXISTS idx_reviews_created_id;

CREATE INDEX IF NOT EXISTS idx_reviews_updated_id
    ON reviews (updated_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_reviews_course_updated
    ON reviews (course_id, updated_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_reviews_course_semester_updated
    ON reviews (course_id, semester, updated_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_reviews_course_rating_updated
    ON reviews (course_id, rating, updated_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_reviews_user_updated
    ON reviews (user_id, updated_at DESC, id DESC);
