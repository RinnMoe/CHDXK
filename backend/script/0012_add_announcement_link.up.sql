ALTER TABLE announcements
    ADD COLUMN IF NOT EXISTS link_url TEXT,
    ADD COLUMN IF NOT EXISTS link_title TEXT;
