CREATE TABLE teacher_profile_sources (url text PRIMARY KEY, data jsonb NOT NULL);
ALTER TABLE teachers ADD COLUMN profile jsonb NOT NULL DEFAULT '{}';
ALTER TABLE teachers ADD COLUMN profile_candidates jsonb NOT NULL DEFAULT '[]';
ALTER TABLE teachers ADD COLUMN profile_status text NOT NULL DEFAULT 'unmatched';
ALTER TABLE teachers ADD COLUMN profile_selected_url text NOT NULL DEFAULT '';
