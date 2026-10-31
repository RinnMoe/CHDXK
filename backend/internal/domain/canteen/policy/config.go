package policy

import "time"

// Config controls canteen review frequency checks. It mirrors the course
// policy shape without coupling the two domains.
type Config struct {
	Window          time.Duration
	MaxReviews      int
	SimilarityRatio float64
}

var DefaultConfig = Config{
	Window:          10 * time.Minute,
	MaxReviews:      5,
	SimilarityRatio: 0.85,
}

func (c Config) Normalized() Config {
	if c.Window <= 0 {
		c.Window = DefaultConfig.Window
	}
	if c.MaxReviews <= 0 {
		c.MaxReviews = DefaultConfig.MaxReviews
	}
	if c.SimilarityRatio <= 0 {
		c.SimilarityRatio = DefaultConfig.SimilarityRatio
	}
	return c
}
