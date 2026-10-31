package canteen

import (
	"strings"
	"time"
	"unicode/utf8"

	"jcourse/pkg/apperr"
)

const (
	MaxReviewContentRunes = 2000
	MaxReviewTags         = 5
)

type ReviewStatus string

const (
	ReviewStatusActive  ReviewStatus = "active"
	ReviewStatusHidden  ReviewStatus = "hidden"
	ReviewStatusDeleted ReviewStatus = "deleted"
)

func (s ReviewStatus) Valid() bool {
	switch s {
	case ReviewStatusActive, ReviewStatusHidden, ReviewStatusDeleted:
		return true
	default:
		return false
	}
}

func (s ReviewStatus) CountedInAggregates() bool {
	return s == ReviewStatusActive
}

// Dimensions holds the five required store scores.
type Dimensions struct {
	Taste   int
	Value   int
	Queue   int
	Hygiene int
	Service int
}

func (d Dimensions) Validate() error {
	for _, value := range d.Values() {
		if value < 1 || value > 5 {
			return ErrRatingOutOfRange
		}
	}
	return nil
}

func (d Dimensions) Values() [5]int {
	return [5]int{d.Taste, d.Value, d.Queue, d.Hygiene, d.Service}
}

func (d Dimensions) Average() DimensionAverages {
	return DimensionAverages{
		Taste:   float64(d.Taste),
		Value:   float64(d.Value),
		Queue:   float64(d.Queue),
		Hygiene: float64(d.Hygiene),
		Service: float64(d.Service),
	}
}

// DishDimensions holds the three required dish scores.
type DishDimensions struct {
	Taste   int
	Portion int
	Value   int
}

func (d DishDimensions) Validate() error {
	for _, value := range [...]int{d.Taste, d.Portion, d.Value} {
		if value < 1 || value > 5 {
			return ErrRatingOutOfRange
		}
	}
	return nil
}

// AllowedTags is the controlled vocabulary rendered by the review form. Tags
// are advisory labels and never take part in ranking.
var AllowedTags = []string{
	"分量足",
	"分量偏少",
	"上菜快",
	"等位久",
	"环境干净",
	"环境嘈杂",
	"服务热情",
	"服务冷淡",
	"性价比高",
	"价格偏高",
	"口味推荐",
	"口味一般",
}

func IsAllowedTag(tag string) bool {
	for _, allowed := range AllowedTags {
		if allowed == tag {
			return true
		}
	}
	return false
}

// ValidateTags checks the controlled vocabulary, the count cap and duplicates.
func ValidateTags(tags []string) error {
	if len(tags) > MaxReviewTags {
		return ErrTooManyTags
	}
	seen := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		if !IsAllowedTag(tag) {
			return ErrInvalidTag
		}
		if _, ok := seen[tag]; ok {
			return ErrDuplicateTag
		}
		seen[tag] = struct{}{}
	}
	return nil
}

func normalizeTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(tags))
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		result = append(result, tag)
	}
	return result
}

// ReviewEvent is one real visit. A user may create many events per store; the
// per-user summary keeps the latest three.
type ReviewEvent struct {
	ID        int
	StoreID   int
	UserID    int
	VisitedAt time.Time
	Dimensions
	OverallScore        float64
	ScoreFormulaVersion int
	Content             string
	Tags                []string
	Status              ReviewStatus
	LikeCount           int
	DislikeCount        int
	ClientRequestID     string
	HiddenReason        string
	HiddenBy            *int
	HiddenAt            *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func validateReviewInput(dimensions Dimensions, content string, tags []string) error {
	if err := dimensions.Validate(); err != nil {
		return err
	}
	if utf8.RuneCountInString(strings.TrimSpace(content)) > MaxReviewContentRunes {
		return ErrContentTooLong
	}
	return ValidateTags(normalizeTags(tags))
}

func (r *ReviewEvent) ApplyCreate(weights DimensionWeights) {
	r.Content = strings.TrimSpace(r.Content)
	r.Tags = normalizeTags(r.Tags)
	r.OverallScore = weights.Score(r.Dimensions)
	r.ScoreFormulaVersion = ScoreFormulaVersion
	if r.Status == "" {
		r.Status = ReviewStatusActive
	}
}

func (r *ReviewEvent) Validate() error {
	if r.StoreID <= 0 || r.UserID <= 0 {
		return apperr.BadRequest("评价归属无效")
	}
	if !r.Status.Valid() {
		return apperr.BadRequest("评价状态无效")
	}
	return validateReviewInput(r.Dimensions, r.Content, r.Tags)
}

// UpdateReview is the editable subset of a visit.
type UpdateReview struct {
	VisitedAt  time.Time
	Dimensions Dimensions
	Content    string
	Tags       []string
	Now        time.Time
}

// ApplyUpdate mutates the event and recomputes the weighted visit score. The
// revision snapshot must be written by the repository in the same transaction.
func (r *ReviewEvent) ApplyUpdate(cmd UpdateReview, weights DimensionWeights) error {
	if err := validateReviewInput(cmd.Dimensions, cmd.Content, cmd.Tags); err != nil {
		return err
	}
	if cmd.VisitedAt.IsZero() {
		cmd.VisitedAt = r.VisitedAt
	}
	r.VisitedAt = cmd.VisitedAt
	r.Dimensions = cmd.Dimensions
	r.Content = strings.TrimSpace(cmd.Content)
	r.Tags = normalizeTags(cmd.Tags)
	r.OverallScore = weights.Score(cmd.Dimensions)
	r.ScoreFormulaVersion = ScoreFormulaVersion
	r.UpdatedAt = cmd.Now
	return nil
}

// Revision stores the pre-update snapshot.
type Revision struct {
	ID            int
	ReviewEventID int
	CreatedBy     int
	VisitedAt     time.Time
	Dimensions    Dimensions
	OverallScore  float64
	Content       string
	Tags          []string
	CreatedAt     time.Time
}

func (r *ReviewEvent) MakeRevision(createdBy int, now time.Time) Revision {
	return Revision{
		ReviewEventID: r.ID,
		CreatedBy:     createdBy,
		VisitedAt:     r.VisitedAt,
		Dimensions:    r.Dimensions,
		OverallScore:  r.OverallScore,
		Content:       r.Content,
		Tags:          append([]string(nil), r.Tags...),
		CreatedAt:     now,
	}
}

type VoteType int

const (
	VoteDislike VoteType = -1
	VoteLike    VoteType = 1
)

func (t VoteType) Valid() bool {
	return t == VoteLike || t == VoteDislike
}

type Vote struct {
	ReviewEventID int
	UserID        int
	VoteType      VoteType
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ReviewListFilter drives the public review list.
type ReviewListFilter struct {
	StoreID   int
	UserID    int
	ViewerID  int
	Status    ReviewStatus
	Order     string
	WithMedia bool
	// History includes every visit. When false the store feed keeps only each
	// user's latest visit so the feed is not dominated by repeat visitors.
	History  bool
	MinScore *float64
	MaxScore *float64
	Cursor   *ReviewCursor
	Page     int
	PageSize int
}

// ReviewCursor is the stable keyset position for a review feed page.
type ReviewCursor struct {
	Order        string    `json:"order"`
	OverallScore float64   `json:"overall_score"`
	LikeCount    int       `json:"like_count"`
	CreatedAt    time.Time `json:"created_at"`
	ID           int       `json:"id"`
}
