CREATE TABLE IF NOT EXISTS point_rewards
(
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER     NOT NULL,
    reason      TEXT        NOT NULL,
    amount      INTEGER     NOT NULL CHECK (amount > 0),
    source_type TEXT        NOT NULL,
    source_key  TEXT        NOT NULL,
    description TEXT        NOT NULL,
    status      TEXT        NOT NULL CHECK (status IN ('pending', 'granted', 'canceled')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    granted_at  TIMESTAMPTZ,

    CONSTRAINT fk_point_rewards_user
        FOREIGN KEY (user_id) REFERENCES users (id)
            ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_point_rewards_reason_source
    ON point_rewards (reason, source_type, source_key);
CREATE INDEX IF NOT EXISTS idx_point_rewards_status_created
    ON point_rewards (status, created_at, id);
CREATE INDEX IF NOT EXISTS idx_point_rewards_user_created
    ON point_rewards (user_id, created_at DESC, id DESC);
