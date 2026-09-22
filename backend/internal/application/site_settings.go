package application

import (
	"context"

	"jcourse/internal/domain/course"
	"jcourse/internal/domain/point"
	"jcourse/internal/domain/review"
	"jcourse/internal/domain/review/policy"
	"jcourse/internal/domain/setting"
)

type SiteSettingsProvider interface {
	ReviewRuntimeConfig(ctx context.Context) (ReviewRuntimeConfig, error)
	HotScoreConfig(ctx context.Context) (course.HotScoreConfig, error)
}

type ReviewRuntimeConfig struct {
	Vote                          review.VoteConfig
	Rewards                       point.RewardConfig
	FrequencyPolicy               policy.FrequencyPolicyConfig
	FrequencyViolationAdminEmails []string
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
			CourseFirstReviewEnabled: snapshot.Bool(setting.SystemSettingKeyReviewRewardCourseFirstReviewEnabled),
			CourseFirstReviewPoints:  snapshot.Int(setting.SystemSettingKeyReviewRewardCourseFirstReviewPts),
			ReviewCreateEnabled:      snapshot.Bool(setting.SystemSettingKeyReviewRewardReviewCreateEnabled),
			ReviewCreatePoints:       snapshot.Int(setting.SystemSettingKeyReviewRewardReviewCreatePts),
		},
		FrequencyPolicy: policy.FrequencyPolicyConfig{
			Window:          snapshot.Duration(setting.SystemSettingKeyReviewFrequencyWindow),
			MaxReviews:      snapshot.Int(setting.SystemSettingKeyReviewFrequencyMaxReviews),
			SimilarityRatio: snapshot.Float(setting.SystemSettingKeyReviewFrequencySimilarityRatio),
			SuspendDuration: snapshot.Duration(setting.SystemSettingKeyReviewFrequencySuspendDuration),
		},
		FrequencyViolationAdminEmails: snapshot.StringList(setting.SystemSettingKeyReviewFrequencyViolationAdminEmails),
	}, nil
}

func (p *SystemSiteSettingsProvider) HotScoreConfig(ctx context.Context) (course.HotScoreConfig, error) {
	snapshot, err := p.settings.Snapshot(ctx)
	if err != nil {
		return course.HotScoreConfig{}, err
	}
	return course.HotScoreConfig{
		ReviewCreateScore: int64(snapshot.Int(setting.SystemSettingKeyReviewHotScoreReviewCreate)),
		ReviewUpdateScore: int64(snapshot.Int(setting.SystemSettingKeyReviewHotScoreReviewUpdate)),
		ReviewVoteScore:   int64(snapshot.Int(setting.SystemSettingKeyReviewHotScoreReviewVote)),
	}, nil
}

type StaticSiteSettingsProvider struct {
	ReviewRuntime ReviewRuntimeConfig
	HotScores     course.HotScoreConfig
}

func NewDefaultSiteSettingsProvider() StaticSiteSettingsProvider {
	return StaticSiteSettingsProvider{
		ReviewRuntime: ReviewRuntimeConfig{
			Vote:                          review.DefaultVoteConfig,
			Rewards:                       point.DefaultRewardConfig,
			FrequencyPolicy:               policy.DefaultFrequencyPolicyConfig,
			FrequencyViolationAdminEmails: []string{},
		},
		HotScores: course.DefaultHotScoreConfig,
	}
}

func (p StaticSiteSettingsProvider) ReviewRuntimeConfig(context.Context) (ReviewRuntimeConfig, error) {
	return p.ReviewRuntime, nil
}

func (p StaticSiteSettingsProvider) HotScoreConfig(context.Context) (course.HotScoreConfig, error) {
	return p.HotScores, nil
}

var _ SiteSettingsProvider = (*SystemSiteSettingsProvider)(nil)
var _ SiteSettingsProvider = StaticSiteSettingsProvider{}
