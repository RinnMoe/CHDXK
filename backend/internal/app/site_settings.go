package app

import (
	"context"
	"strconv"
	"strings"
	"time"

	"jcourse/internal/domain/account"
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
			Key:          setting.SystemSettingKeyAuthEmailDomain,
			Group:        "auth",
			Label:        "邮箱后缀",
			Description:  "登录、注册和重置密码表单默认拼接的邮箱后缀",
			Type:         setting.ValueTypeString,
			DefaultValue: "@sjtu.edu.cn",
			Public:       true,
			Validator:    setting.ValueValidatorFunc(emailDomainSetting),
		},
		{
			Key:          setting.SystemSettingKeyAuthRegistrationEmailWhitelist,
			Group:        "auth",
			Label:        "注册邮箱白名单",
			Description:  "允许注册的邮箱或邮箱后缀，多个值用逗号或换行分隔",
			Type:         setting.ValueTypeStringList,
			DefaultValue: "@sjtu.edu.cn",
			Validator:    setting.ValueValidatorFunc(nonEmptyStringListSetting),
		},
		{
			Key:          setting.SystemSettingKeyAuthLoginMaxAttempts,
			Group:        "auth",
			Label:        "登录失败次数上限",
			Description:  "同一邮箱达到该失败次数后会被临时锁定",
			Type:         setting.ValueTypeInt,
			DefaultValue: strconv.Itoa(account.DefaultLoginConfig.MaxAttempts),
			Validator:    setting.ValueValidatorFunc(positiveIntSetting),
		},
		{
			Key:          setting.SystemSettingKeyAuthLoginLockout,
			Group:        "auth",
			Label:        "登录锁定时长",
			Description:  "登录失败次数达到上限后的锁定时长",
			Type:         setting.ValueTypeDuration,
			DefaultValue: account.DefaultLoginConfig.Lockout.String(),
			Validator:    setting.ValueValidatorFunc(positiveDurationSetting),
		},
		{
			Key:          setting.SystemSettingKeyAuthVerificationCodeInterval,
			Group:        "auth",
			Label:        "验证码发送间隔",
			Description:  "同一邮箱重复发送验证码的最短间隔",
			Type:         setting.ValueTypeDuration,
			DefaultValue: "1m",
			Validator:    setting.ValueValidatorFunc(positiveDurationSetting),
		},
		{
			Key:          setting.SystemSettingKeyAuthVerificationCodeTTL,
			Group:        "auth",
			Label:        "验证码有效期",
			Description:  "注册和重置密码验证码的有效时间",
			Type:         setting.ValueTypeDuration,
			DefaultValue: "10m",
			Validator:    setting.ValueValidatorFunc(positiveDurationSetting),
		},
		{
			Key:          setting.SystemSettingKeyAPIKeyMaxUserKeys,
			Group:        "api_key",
			Label:        "用户 API key 数量上限",
			Description:  "单个用户最多可创建的 API key 数量",
			Type:         setting.ValueTypeInt,
			DefaultValue: strconv.Itoa(auth.DefaultApiKeyConfig.MaxUserKeys),
			Validator:    setting.ValueValidatorFunc(positiveIntSetting),
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
			Key:          setting.SystemSettingKeyReviewRewardCourseFirstReviewEnabled,
			Group:        "review_reward",
			Label:        "启用首评奖励",
			Description:  "是否启用课程首评积分奖励",
			Type:         setting.ValueTypeBool,
			DefaultValue: strconv.FormatBool(point.DefaultRewardConfig.CourseFirstReviewEnabled),
		},
		{
			Key:          setting.SystemSettingKeyReviewRewardCourseFirstReviewPts,
			Group:        "review_reward",
			Label:        "课程首评奖励积分",
			Description:  "课程首次点评奖励的积分数量",
			Type:         setting.ValueTypeInt,
			DefaultValue: strconv.Itoa(point.DefaultRewardConfig.CourseFirstReviewPoints),
			Validator:    setting.ValueValidatorFunc(nonNegativeIntSetting),
		},
		{
			Key:          setting.SystemSettingKeyReviewRewardReviewCreateEnabled,
			Group:        "review_reward",
			Label:        "启用发布点评奖励",
			Description:  "是否启用每条新点评的积分奖励",
			Type:         setting.ValueTypeBool,
			DefaultValue: strconv.FormatBool(point.DefaultRewardConfig.ReviewCreateEnabled),
		},
		{
			Key:          setting.SystemSettingKeyReviewRewardReviewCreatePts,
			Group:        "review_reward",
			Label:        "发布点评奖励积分",
			Description:  "每条新点评奖励的积分数量",
			Type:         setting.ValueTypeInt,
			DefaultValue: strconv.Itoa(point.DefaultRewardConfig.ReviewCreatePoints),
			Validator:    setting.ValueValidatorFunc(nonNegativeIntSetting),
		},
		{
			Key:          setting.SystemSettingKeyReviewHotScoreReviewCreate,
			Group:        "review",
			Label:        "发布点评热度分",
			Description:  "发布点评时给课程热度增加的分值",
			Type:         setting.ValueTypeInt,
			DefaultValue: strconv.FormatInt(course.DefaultHotScoreConfig.ReviewCreateScore, 10),
			Validator:    setting.ValueValidatorFunc(nonNegativeIntSetting),
		},
		{
			Key:          setting.SystemSettingKeyReviewHotScoreReviewUpdate,
			Group:        "review",
			Label:        "更新点评热度分",
			Description:  "更新点评时给课程热度增加的分值",
			Type:         setting.ValueTypeInt,
			DefaultValue: strconv.FormatInt(course.DefaultHotScoreConfig.ReviewUpdateScore, 10),
			Validator:    setting.ValueValidatorFunc(nonNegativeIntSetting),
		},
		{
			Key:          setting.SystemSettingKeyReviewHotScoreReviewVote,
			Group:        "review",
			Label:        "点评投票热度分",
			Description:  "点评获得投票时给课程热度增加的分值",
			Type:         setting.ValueTypeInt,
			DefaultValue: strconv.FormatInt(course.DefaultHotScoreConfig.ReviewVoteScore, 10),
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
			Key:          setting.SystemSettingKeyReviewFrequencyViolationAdminEmails,
			Group:        "review",
			Label:        "刷评通知管理员邮箱",
			Description:  "命中刷评策略后接收通知的管理员邮箱，多个值用逗号或换行分隔，留空则不发送",
			Type:         setting.ValueTypeStringList,
			DefaultValue: "",
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

func emailDomainSetting(_ context.Context, value string) error {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "@") || len(value) <= 1 || strings.Contains(value[1:], "@") {
		return setting.ErrInvalidSystemSettingValue
	}
	return nil
}

func nonEmptyStringListSetting(_ context.Context, value string) error {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r' || r == '\t'
	})
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			return nil
		}
	}
	return setting.ErrInvalidSystemSettingValue
}
