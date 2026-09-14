-- Keep an audit/rollback copy of the exact legacy defaults before changing them.
CREATE TABLE IF NOT EXISTS system_settings_localization_backup_0015
(
    key         TEXT PRIMARY KEY,
    value       TEXT        NOT NULL,
    backed_up_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO system_settings_localization_backup_0015 (key, value)
SELECT key, value
FROM system_settings
WHERE key IN ('auth.email_domain', 'auth.registration.email_whitelist')
  AND value = '@sjtu.edu.cn'
ON CONFLICT (key) DO NOTHING;

UPDATE system_settings
SET value      = '@chd.edu.cn',
    updated_at = NOW()
WHERE key IN ('auth.email_domain', 'auth.registration.email_whitelist')
  AND value = '@sjtu.edu.cn';
