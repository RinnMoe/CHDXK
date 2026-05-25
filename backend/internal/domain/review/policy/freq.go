package policy

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/agnivade/levenshtein"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	domainemail "jcourse/internal/domain/email"
	"jcourse/internal/domain/review"
	"jcourse/internal/domain/task"
)

var (
	ErrSimilarContentDetected = errors.New("too many similar reviews detected")
	ErrSameCourseSpam         = errors.New("too many reviews for the same course")
)

type FrequencyPolicyConfig struct {
	Window          time.Duration
	MaxReviews      int
	SimilarityRatio float64
	SuspendDuration time.Duration
	AdminEmails     []string
}

var DefaultFrequencyPolicyConfig = FrequencyPolicyConfig{
	Window:          time.Hour,
	MaxReviews:      10,
	SimilarityRatio: 0.7,
	SuspendDuration: 90 * 24 * time.Hour,
}

const (
	spamSuspensionEmailType     = "account_banned"
	spamSuspensionEmailSubject  = "选课社区用户封禁通知"
	spamSuspensionEmailTemplate = `选课社区用户封禁通知

管理员您好：

系统检测到用户触发点评频率策略，已创建封禁任务。

用户 ID：{{.UserID}}
课程代码：{{.CourseCode}}
封禁时长：{{.Duration}}
原因：{{.Reason}}

请在后台查看用户和点评记录。

选课社区
`
)

type FrequencyPolicy struct {
	query  review.ReviewQuery
	config FrequencyPolicyConfig
}

func NewFrequencyPolicy(query review.ReviewQuery, config FrequencyPolicyConfig) *FrequencyPolicy {
	defaults := DefaultFrequencyPolicyConfig
	if config.Window <= 0 {
		config.Window = defaults.Window
	}
	if config.MaxReviews <= 0 {
		config.MaxReviews = defaults.MaxReviews
	}
	if config.SimilarityRatio <= 0 {
		config.SimilarityRatio = defaults.SimilarityRatio
	}
	if config.SuspendDuration <= 0 {
		config.SuspendDuration = defaults.SuspendDuration
	}
	return &FrequencyPolicy{query: query, config: config}
}

func (p *FrequencyPolicy) CanCreate(ctx context.Context, u *auth.User, c *course.Course, r *review.Review) error {
	if p.config.MaxReviews <= 0 {
		return nil
	}

	recent, _, err := p.query.FindBy(ctx, review.ReviewFilter{
		UserID:       u.ID,
		CreatedAfter: time.Now().Add(-p.config.Window),
		OrderBy:      "created_at",
		PageSize:     p.config.MaxReviews,
		WithCourse:   true,
	})
	if err != nil {
		return err
	}

	if len(recent) < p.config.MaxReviews {
		return nil
	}

	similarCount := 0
	sameCourseCodeAll := true
	for _, rev := range recent {
		if !reviewCourseMatches(rev, c) {
			sameCourseCodeAll = false
		}
		if similarity(r.Content, rev.Content) >= p.config.SimilarityRatio {
			similarCount++
		}
	}

	if sameCourseCodeAll {
		p.enqueueSuspensionTasks(ctx, u.ID, c, ErrSameCourseSpam)
		return ErrSameCourseSpam
	}

	if similarCount*2 > len(recent) {
		p.enqueueSuspensionTasks(ctx, u.ID, c, ErrSimilarContentDetected)
		return ErrSimilarContentDetected
	}

	return nil
}

func (p *FrequencyPolicy) enqueueSuspensionTasks(ctx context.Context, userID int, c *course.Course, reason error) {
	_ = task.Enqueue(ctx, auth.NewSuspendUserTask(userID, p.config.SuspendDuration))
	for _, to := range p.config.AdminEmails {
		_ = task.Enqueue(ctx, domainemail.NewSendEmailTask(
			spamSuspensionEmailType,
			to,
			spamSuspensionEmailSubject,
			spamSuspensionEmailTemplate,
			map[string]string{
				"UserID":     fmt.Sprint(userID),
				"CourseCode": c.Code,
				"Duration":   formatSuspensionDuration(p.config.SuspendDuration),
				"Reason":     reason.Error(),
			},
		))
	}
}

func formatSuspensionDuration(d time.Duration) string {
	if d <= 0 {
		return "一段时间"
	}
	if d%(24*time.Hour) == 0 {
		return fmt.Sprintf("%d 天", int(d/(24*time.Hour)))
	}
	if d%time.Hour == 0 {
		return fmt.Sprintf("%d 小时", int(d/time.Hour))
	}
	return d.String()
}

func reviewCourseMatches(r review.ReviewView, c *course.Course) bool {
	if r.Course != nil && c.Code != "" {
		return r.Course.Code == c.Code
	}
	return r.CourseID == c.ID
}

func similarity(a, b string) float64 {
	if a == "" && b == "" {
		return 1.0
	}
	maxLen := max(len(a), len(b))
	if maxLen == 0 {
		return 1.0
	}
	dist := levenshtein.ComputeDistance(a, b)
	return 1.0 - float64(dist)/float64(maxLen)
}
