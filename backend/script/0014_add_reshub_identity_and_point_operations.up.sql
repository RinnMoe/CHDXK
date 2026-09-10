CREATE TABLE IF NOT EXISTS user_identities
(
    user_id       INTEGER     PRIMARY KEY,
    issuer        TEXT        NOT NULL,
    subject       TEXT        NOT NULL,
    username      TEXT        NOT NULL DEFAULT '',
    email         TEXT        NOT NULL DEFAULT '',
    identity_role TEXT        NOT NULL DEFAULT '',
    identity_type TEXT        NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_user_identities_user
        FOREIGN KEY (user_id) REFERENCES users (id)
            ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_user_identities_issuer_subject
    ON user_identities (issuer, subject);

ALTER TABLE point_rewards
    ADD COLUMN IF NOT EXISTS operation_key TEXT,
    ADD COLUMN IF NOT EXISTS external_operation_id TEXT,
    ADD COLUMN IF NOT EXISTS external_reversal_id TEXT,
    ADD COLUMN IF NOT EXISTS attempts INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS last_error TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMPTZ;

CREATE UNIQUE INDEX IF NOT EXISTS uniq_point_rewards_operation_key
    ON point_rewards (operation_key)
    WHERE operation_key IS NOT NULL AND operation_key <> '';

CREATE INDEX IF NOT EXISTS idx_point_rewards_next_retry
    ON point_rewards (next_retry_at, id)
    WHERE next_retry_at IS NOT NULL;
