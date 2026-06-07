package app

import (
	"context"
	"strconv"
	"time"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/point"
	"jcourse/internal/domain/review"
	"jcourse/internal/domain/review/policy"
	"jcourse/internal/domain/setting"
)

func newSystemSettingsService(repo setting.SystemRepository, courseRepo course.CourseRepository) *setting.SystemSettingsService {
	registry := setting.NewRegistry([]setting.Definition{
		{
			Key:          setting.SystemSettingKeyCurrentSemester,
			Group:        "course",
			Label:        "当前学期",
			Description:  "前台课程和写评默认使用的当前学期",
			Type:         setting.ValueTypeString,
			DefaultValue: "",
			Public:       true,
			Validator:    setting.NewCurrentSemesterValidator(courseRepo),
		},
		{
			Key:          setting.SystemSettingKeyReviewVoteMaxDailyVotes,
			Group:        "review",
			Label:        "每日投票上限",
			Description:  "单个用户每天可投票的最大次数",
			Type:         setting.ValueTypeInt,
			DefaultValue: strconv.Itoa(review.DefaultVoteConfig.MaxDailyVotes),
			Validator:    setting.ValueValidatorFunc(positiveIntSetting),
		},
		{
			Key:          setting.SystemSettingKeyReviewRewardEnabled,
			Group:        "review",
			Label:        "启用首评奖励",
			Description:  "是否启用课程首评积分奖励",
			Type:         setting.ValueTypeBool,
			DefaultValue: strconv.FormatBool(point.DefaultRewardConfig.Enabled),
		},
		{
			Key:          setting.SystemSettingKeyReviewRewardCourseFirstReviewPts,
			Group:        "review",
			Label:        "课程首评奖励积分",
			Description:  "课程首次点评奖励的积分数量",
			Type:         setting.ValueTypeInt,
			DefaultValue: strconv.Itoa(point.DefaultRewardConfig.CourseFirstReviewPoints),
			Validator:    setting.ValueValidatorFunc(nonNegativeIntSetting),
		},
		{
			Key:          setting.SystemSettingKeyReviewFrequencyWindow,
			Group:        "review",
			Label:        "刷评检测窗口",
			Description:  "检测短时间内刷评的时间窗口",
			Type:         setting.ValueTypeDuration,
			DefaultValue: policy.DefaultFrequencyPolicyConfig.Window.String(),
			Validator:    setting.ValueValidatorFunc(positiveDurationSetting),
		},
		{
			Key:          setting.SystemSettingKeyReviewFrequencyMaxReviews,
			Group:        "review",
			Label:        "刷评检测评论数阈值",
			Description:  "进入刷评判定前需要命中的最少评论数",
			Type:         setting.ValueTypeInt,
			DefaultValue: strconv.Itoa(policy.DefaultFrequencyPolicyConfig.MaxReviews),
			Validator:    setting.ValueValidatorFunc(positiveIntSetting),
		},
		{
			Key:          setting.SystemSettingKeyReviewFrequencySimilarityRatio,
			Group:        "review",
			Label:        "刷评相似度阈值",
			Description:  "评论文本相似度达到该值时视为相似",
			Type:         setting.ValueTypeFloat,
			DefaultValue: strconv.FormatFloat(policy.DefaultFrequencyPolicyConfig.SimilarityRatio, 'f', -1, 64),
			Validator:    setting.ValueValidatorFunc(ratioSetting),
		},
		{
			Key:          setting.SystemSettingKeyReviewFrequencySuspendDuration,
			Group:        "review",
			Label:        "刷评封禁时长",
			Description:  "命中刷评策略后的封禁时长",
			Type:         setting.ValueTypeDuration,
			DefaultValue: policy.DefaultFrequencyPolicyConfig.SuspendDuration.String(),
			Validator:    setting.ValueValidatorFunc(positiveDurationSetting),
		},
		{
			Key:          setting.SystemSettingKeyAdminDefaultSuspendDays,
			Group:        "admin",
			Label:        "默认封禁天数",
			Description:  "管理员未填写时的默认封禁天数",
			Type:         setting.ValueTypeInt,
			DefaultValue: strconv.Itoa(auth.DefaultAdminConfig.DefaultSuspendDays),
			Validator:    setting.ValueValidatorFunc(positiveIntSetting),
		},
	})
	return setting.NewSystemSettingsService(repo, registry)
}

func positiveIntSetting(_ context.Context, value string) error {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return setting.ErrInvalidSystemSettingValue
	}
	return nil
}

func nonNegativeIntSetting(_ context.Context, value string) error {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return setting.ErrInvalidSystemSettingValue
	}
	return nil
}

func positiveDurationSetting(_ context.Context, value string) error {
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return setting.ErrInvalidSystemSettingValue
	}
	return nil
}

func ratioSetting(_ context.Context, value string) error {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || parsed <= 0 || parsed > 1 {
		return setting.ErrInvalidSystemSettingValue
	}
	return nil
}
