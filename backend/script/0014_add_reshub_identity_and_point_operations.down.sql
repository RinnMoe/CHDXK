DROP INDEX IF EXISTS idx_point_rewards_next_retry;
DROP INDEX IF EXISTS uniq_point_rewards_operation_key;

ALTER TABLE point_rewards
    DROP COLUMN IF EXISTS next_retry_at,
    DROP COLUMN IF EXISTS last_error,
    DROP COLUMN IF EXISTS attempts,
    DROP COLUMN IF EXISTS external_reversal_id,
    DROP COLUMN IF EXISTS external_operation_id,
    DROP COLUMN IF EXISTS operation_key;

DROP INDEX IF EXISTS uniq_user_identities_issuer_subject;
DROP TABLE IF EXISTS user_identities;
