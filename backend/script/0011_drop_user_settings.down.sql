CREATE TABLE IF NOT EXISTS user_settings
(
    user_id          INTEGER     PRIMARY KEY,
    current_semester TEXT        NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_user_settings_user
        FOREIGN KEY (user_id) REFERENCES users (id)
            ON DELETE CASCADE
);
