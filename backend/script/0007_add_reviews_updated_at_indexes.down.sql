DROP INDEX IF EXISTS idx_reviews_user_updated;
DROP INDEX IF EXISTS idx_reviews_course_rating_updated;
DROP INDEX IF EXISTS idx_reviews_course_semester_updated;
DROP INDEX IF EXISTS idx_reviews_course_updated;
DROP INDEX IF EXISTS idx_reviews_updated_id;

CREATE INDEX IF NOT EXISTS idx_reviews_created_id
    ON reviews (created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_reviews_course_created
    ON reviews (course_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_reviews_course_semester_created
    ON reviews (course_id, semester, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_reviews_course_rating_created
    ON reviews (course_id, rating, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_reviews_user_id_desc
    ON reviews (user_id, id DESC);
