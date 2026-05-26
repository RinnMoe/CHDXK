BEGIN;

ALTER TABLE review_revisions
    RENAME COLUMN user_id TO created_by; -- PostgreSQL has no IF EXISTS for RENAME COLUMN before 17.

CREATE INDEX IF NOT EXISTS idx_review_revisions_created_by
    ON review_revisions (created_by, created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS audit_logs
(
    id            BIGSERIAL PRIMARY KEY,
    occurred_at   TIMESTAMPTZ NOT NULL,
    actor_user_id INTEGER     NOT NULL,
    action        TEXT        NOT NULL,
    target_type   TEXT        NOT NULL,
    target_id     TEXT        NOT NULL,
    details       JSONB       NOT NULL DEFAULT '{}'::jsonb,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_occurred
    ON audit_logs (occurred_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_occurred
    ON audit_logs (actor_user_id, occurred_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action_occurred
    ON audit_logs (action, occurred_at DESC, id DESC);

COMMIT;
