ALTER TABLE point_rewards DROP COLUMN external_subject;
DROP TABLE bootstrap_accounts;
-- Existing accounts retain their IDs to preserve references and audit history.
