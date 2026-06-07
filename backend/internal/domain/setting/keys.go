package setting

const (
	SystemSettingKeyCurrentSemester = "current_semester"
	SystemSettingKeyAuthEmailDomain = "auth.email_domain"

	SystemSettingKeyAuthRegistrationEmailWhitelist = "auth.registration.email_whitelist"
	SystemSettingKeyAuthLoginMaxAttempts           = "auth.login.max_attempts"
	SystemSettingKeyAuthLoginLockout               = "auth.login.lockout"
	SystemSettingKeyAuthVerificationCodeInterval   = "auth.verification.code_interval"
	SystemSettingKeyAuthVerificationCodeTTL        = "auth.verification.code_ttl"

	SystemSettingKeyAPIKeyMaxUserKeys = "api_key.max_user_keys"

	SystemSettingKeyReviewVoteMaxDailyVotes          = "review.vote.max_daily_votes"
	SystemSettingKeyReviewRewardEnabled              = "review.rewards.enabled"
	SystemSettingKeyReviewRewardCourseFirstReviewPts = "review.rewards.course_first_review_points"
	SystemSettingKeyReviewHotScoreReviewCreate       = "review.hot_scores.review_create_score"
	SystemSettingKeyReviewHotScoreReviewUpdate       = "review.hot_scores.review_update_score"
	SystemSettingKeyReviewHotScoreReviewVote         = "review.hot_scores.review_vote_score"

	SystemSettingKeyReviewFrequencyWindow               = "review.frequency.window"
	SystemSettingKeyReviewFrequencyMaxReviews           = "review.frequency.max_reviews"
	SystemSettingKeyReviewFrequencySimilarityRatio      = "review.frequency.similarity_ratio"
	SystemSettingKeyReviewFrequencySuspendDuration      = "review.frequency.suspend_duration"
	SystemSettingKeyReviewFrequencyViolationAdminEmails = "review.frequency.violation_admin_emails"
)
