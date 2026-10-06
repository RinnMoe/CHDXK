ALTER TABLE users
  ADD COLUMN nickname TEXT NOT NULL DEFAULT '',
  ADD COLUMN must_change_password BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN auth_version INTEGER NOT NULL DEFAULT 1,
  ADD COLUMN permissions JSONB NOT NULL DEFAULT '[]',
  ADD COLUMN muted_until TIMESTAMPTZ,
  ADD COLUMN moderation_reason TEXT NOT NULL DEFAULT '',
  ADD COLUMN deleted_at TIMESTAMPTZ;
ALTER TABLE user_identities ADD COLUMN has_points_account BOOLEAN NOT NULL DEFAULT false;
DROP INDEX IF EXISTS uniq_user_identities_issuer_subject;
CREATE INDEX idx_user_identities_subject ON user_identities(issuer, subject);
INSERT INTO user_identities(user_id, issuer, subject, username, email, identity_role, identity_type, created_at, updated_at, has_points_account)
SELECT user_id, issuer, subject, nickname, '', 'user', 'member', NOW(), NOW(), has_points_account
FROM bootstrap_accounts WHERE subject <> '' ON CONFLICT(user_id) DO NOTHING;
UPDATE users SET username='hub_' || id::text WHERE username LIKE 'oidc_%';
UPDATE users u SET nickname=i.username FROM user_identities i WHERE i.user_id=u.id;
-- Local roles and inherited roles are independent. Only the initial owner is super admin.
UPDATE users SET role='user' WHERE id<>1;
UPDATE users SET role='super_admin', nickname='初始超级管理员' WHERE id=1;
DROP TABLE bootstrap_accounts;
