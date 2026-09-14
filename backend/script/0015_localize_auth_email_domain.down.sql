-- Restore only values that still equal the localized default. The backup table
-- is intentionally kept as an audit trail for the migration's original values.
UPDATE system_settings AS settings
SET value      = backup.value,
    updated_at = NOW()
FROM system_settings_localization_backup_0015 AS backup
WHERE settings.key = backup.key
  AND settings.value = '@chd.edu.cn'
  AND backup.value = '@sjtu.edu.cn';
