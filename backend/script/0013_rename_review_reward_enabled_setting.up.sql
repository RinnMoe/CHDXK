UPDATE system_settings
SET key = 'review.rewards.course_first_review_enabled',
    updated_at = NOW()
WHERE key = 'review.rewards.enabled'
  AND NOT EXISTS (
      SELECT 1
      FROM system_settings
      WHERE key = 'review.rewards.course_first_review_enabled'
  );

DELETE FROM system_settings
WHERE key = 'review.rewards.enabled';
