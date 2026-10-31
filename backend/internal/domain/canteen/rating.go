package canteen

import "math"

const (
	// ScoreFormulaVersion is persisted with every review event so a later
	// formula change can be traced without rewriting history.
	ScoreFormulaVersion = 1
	DefaultPriorCount   = 5
	MaxRecentEvents     = 3
)

// DimensionWeights is the store-level weighting of the five required scores.
type DimensionWeights struct {
	Taste   float64
	Value   float64
	Queue   float64
	Hygiene float64
	Service float64
}

var DefaultStoreWeights = DimensionWeights{
	Taste:   0.35,
	Value:   0.20,
	Queue:   0.15,
	Hygiene: 0.20,
	Service: 0.10,
}

// DishWeights is the dish-level weighting of the three required scores.
type DishWeights struct {
	Taste   float64
	Portion float64
	Value   float64
}

var DefaultDishWeights = DishWeights{
	Taste:   0.50,
	Portion: 0.25,
	Value:   0.25,
}

func (w DimensionWeights) Score(d Dimensions) float64 {
	return Round2(
		w.Taste*float64(d.Taste) +
			w.Value*float64(d.Value) +
			w.Queue*float64(d.Queue) +
			w.Hygiene*float64(d.Hygiene) +
			w.Service*float64(d.Service),
	)
}

func (w DishWeights) Score(d DishDimensions) float64 {
	return Round2(
		w.Taste*float64(d.Taste) +
			w.Portion*float64(d.Portion) +
			w.Value*float64(d.Value),
	)
}

var recentEventWeights = [MaxRecentEvents]float64{0.6, 0.3, 0.1}

// RecentWeights returns the weights applied to the most recent events, newest
// first. A user with a single event gets weight 1.0 after normalization.
func RecentWeights(n int) []float64 {
	if n <= 0 {
		return nil
	}
	if n > MaxRecentEvents {
		n = MaxRecentEvents
	}
	weights := make([]float64, n)
	copy(weights, recentEventWeights[:n])
	return weights
}

// WeightedScore computes sum(score*weight)/sum(weight) and therefore handles
// users with fewer than three events without imputing missing scores.
func WeightedScore(scores []float64, weights []float64) float64 {
	limit := len(scores)
	if len(weights) < limit {
		limit = len(weights)
	}
	if limit == 0 {
		return 0
	}
	var numerator, denominator float64
	for i := 0; i < limit; i++ {
		numerator += scores[i] * weights[i]
		denominator += weights[i]
	}
	if denominator == 0 {
		return 0
	}
	return Round2(numerator / denominator)
}

// BayesScore shrinks a store average toward the platform mean. count is the
// number of distinct rating users, not the number of visit events.
func BayesScore(avg float64, count int, globalAvg float64, prior int) float64 {
	if count <= 0 {
		return 0
	}
	if prior <= 0 {
		return Round2(avg)
	}
	return Round2((avg*float64(count) + globalAvg*float64(prior)) / float64(count+prior))
}

func Round2(value float64) float64 {
	return math.Round(value*100) / 100
}

// RatingConfidence turns the deduplicated reviewer count into a user-facing
// confidence label. It describes sample size only, never data quality.
func RatingConfidence(ratingCount int) string {
	switch {
	case ratingCount <= 0:
		return ""
	case ratingCount < 3:
		return "样本较少"
	case ratingCount < 10:
		return "参考"
	default:
		return "较可信"
	}
}

// RatingConfig carries the ranking prior and the periodic refresh schedule.
type RatingConfig struct {
	PriorCount       int    `mapstructure:"prior_count"`
	RefreshCron      string `mapstructure:"refresh_cron"`
	SchedulerEnabled bool   `mapstructure:"scheduler_enabled"`
}

var DefaultRatingConfig = RatingConfig{
	PriorCount:       DefaultPriorCount,
	RefreshCron:      "30 5 * * *",
	SchedulerEnabled: true,
}

func (c RatingConfig) Normalized() RatingConfig {
	if c.PriorCount <= 0 {
		c.PriorCount = DefaultPriorCount
	}
	if c.RefreshCron == "" {
		c.RefreshCron = DefaultRatingConfig.RefreshCron
	}
	return c
}
