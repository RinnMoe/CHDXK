CREATE TABLE bootstrap_accounts (
  user_id INTEGER PRIMARY KEY CHECK(user_id=1) REFERENCES users(id),
  token_hash TEXT NOT NULL DEFAULT '', issuer TEXT NOT NULL DEFAULT '',
  subject TEXT NOT NULL DEFAULT '', nickname TEXT NOT NULL DEFAULT '',
  has_points_account BOOLEAN NOT NULL DEFAULT false
);
-- Multiple bindings cannot be represented by the old model; preserve one per subject.
DELETE FROM user_identities a USING user_identities b
WHERE a.issuer=b.issuer AND a.subject=b.subject AND a.user_id>b.user_id;
DROP INDEX idx_user_identities_subject;
CREATE UNIQUE INDEX uniq_user_identities_issuer_subject ON user_identities(issuer,subject);
ALTER TABLE user_identities DROP COLUMN has_points_account;
ALTER TABLE users DROP COLUMN nickname, DROP COLUMN must_change_password,
  DROP COLUMN auth_version, DROP COLUMN permissions, DROP COLUMN muted_until,
  DROP COLUMN moderation_reason, DROP COLUMN deleted_at;
