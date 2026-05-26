package course

const DefaultRatingScorePriorCount = 5

const TaskTypeRefreshRatingScores = "course:refresh_rating_scores"

type RatingScoreConfig struct {
	PriorCount       int    `mapstructure:"prior_count"`
	RefreshCron      string `mapstructure:"refresh_cron"`
	SchedulerEnabled bool   `mapstructure:"scheduler_enabled"`
}

var DefaultRatingScoreConfig = RatingScoreConfig{
	PriorCount:       DefaultRatingScorePriorCount,
	RefreshCron:      "0 5 * * *",
	SchedulerEnabled: true,
}

func (c RatingScoreConfig) Normalized() RatingScoreConfig {
	if c.PriorCount <= 0 {
		c.PriorCount = DefaultRatingScorePriorCount
	}
	if c.RefreshCron == "" {
		c.RefreshCron = DefaultRatingScoreConfig.RefreshCron
	}
	return c
}

type RefreshRatingScoresPayload struct{}

type RefreshRatingScoresTask struct{}

func NewRefreshRatingScoresTask() RefreshRatingScoresTask {
	return RefreshRatingScoresTask{}
}

func (t RefreshRatingScoresTask) Type() string {
	return TaskTypeRefreshRatingScores
}

func (t RefreshRatingScoresTask) Payload() []byte {
	return []byte("{}")
}
