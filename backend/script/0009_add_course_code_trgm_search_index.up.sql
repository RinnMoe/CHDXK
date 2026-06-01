CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_courses_code_trgm ON courses USING GIN (code gin_trgm_ops);
