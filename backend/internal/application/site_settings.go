package application

import (
	"context"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/point"
	"jcourse/internal/domain/review"
	"jcourse/internal/domain/review/policy"
	"jcourse/internal/domain/setting"
)

type SiteSettingsProvider interface {
	ReviewRuntimeConfig(ctx context.Context) (ReviewRuntimeConfig, error)
	AdminUserConfig(ctx context.Context) (AdminUserCommandConfig, error)
}

type ReviewRuntimeConfig struct {
	Vote            review.VoteConfig
	Rewards         point.RewardConfig
	FrequencyPolicy policy.FrequencyPolicyConfig
}

type SystemSiteSettingsProvider struct {
	settings *setting.SystemSettingsService
}

func NewSystemSiteSettingsProvider(settings *setting.SystemSettingsService) *SystemSiteSettingsProvider {
	return &SystemSiteSettingsProvider{settings: settings}
}

func (p *SystemSiteSettingsProvider) ReviewRuntimeConfig(ctx context.Context) (ReviewRuntimeConfig, error) {
	snapshot, err := p.settings.Snapshot(ctx)
	if err != nil {
		return ReviewRuntimeConfig{}, err
	}
	return ReviewRuntimeConfig{
		Vote: review.VoteConfig{
			MaxDailyVotes: snapshot.Int(setting.SystemSettingKeyReviewVoteMaxDailyVotes),
		},
		Rewards: point.RewardConfig{
			Enabled:                 snapshot.Bool(setting.SystemSettingKeyReviewRewardEnabled),
			CourseFirstReviewPoints: snapshot.Int(setting.SystemSettingKeyReviewRewardCourseFirstReviewPts),
		},
		FrequencyPolicy: policy.FrequencyPolicyConfig{
			Window:          snapshot.Duration(setting.SystemSettingKeyReviewFrequencyWindow),
			MaxReviews:      snapshot.Int(setting.SystemSettingKeyReviewFrequencyMaxReviews),
			SimilarityRatio: snapshot.Float(setting.SystemSettingKeyReviewFrequencySimilarityRatio),
			SuspendDuration: snapshot.Duration(setting.SystemSettingKeyReviewFrequencySuspendDuration),
		},
	}, nil
}

func (p *SystemSiteSettingsProvider) AdminUserConfig(ctx context.Context) (AdminUserCommandConfig, error) {
	snapshot, err := p.settings.Snapshot(ctx)
	if err != nil {
		return AdminUserCommandConfig{}, err
	}
	return AdminUserCommandConfig{
		DefaultSuspendDays: snapshot.Int(setting.SystemSettingKeyAdminDefaultSuspendDays),
	}, nil
}

type StaticSiteSettingsProvider struct {
	ReviewRuntime ReviewRuntimeConfig
	AdminUser     AdminUserCommandConfig
}

func NewDefaultSiteSettingsProvider() StaticSiteSettingsProvider {
	return StaticSiteSettingsProvider{
		ReviewRuntime: ReviewRuntimeConfig{
			Vote:            review.DefaultVoteConfig,
			Rewards:         point.DefaultRewardConfig,
			FrequencyPolicy: policy.DefaultFrequencyPolicyConfig,
		},
		AdminUser: AdminUserCommandConfig{DefaultSuspendDays: auth.DefaultAdminConfig.DefaultSuspendDays},
	}
}

func (p StaticSiteSettingsProvider) ReviewRuntimeConfig(context.Context) (ReviewRuntimeConfig, error) {
	return p.ReviewRuntime, nil
}

func (p StaticSiteSettingsProvider) AdminUserConfig(context.Context) (AdminUserCommandConfig, error) {
	return p.AdminUser, nil
}

var _ SiteSettingsProvider = (*SystemSiteSettingsProvider)(nil)
var _ SiteSettingsProvider = StaticSiteSettingsProvider{}
