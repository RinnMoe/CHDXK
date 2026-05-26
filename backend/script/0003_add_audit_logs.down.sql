DROP INDEX IF EXISTS idx_audit_logs_action_occurred;
DROP INDEX IF EXISTS idx_audit_logs_actor_occurred;
DROP INDEX IF EXISTS idx_audit_logs_occurred;
DROP TABLE IF EXISTS audit_logs;

DROP INDEX IF EXISTS idx_review_revisions_created_by;

ALTER TABLE review_revisions
    RENAME COLUMN created_by TO user_id;
