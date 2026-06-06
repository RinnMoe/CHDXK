CREATE TABLE IF NOT EXISTS system_settings
(
    key        TEXT        PRIMARY KEY,
    value      TEXT        NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO system_settings (key, value)
VALUES ('current_semester', '2025-2026-2')
ON CONFLICT (key) DO NOTHING;
