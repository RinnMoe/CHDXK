package canteen

import (
	"errors"
	"testing"
	"time"
)

func TestDimensionsValidate(t *testing.T) {
	if err := (Dimensions{5, 5, 5, 5, 5}).Validate(); err != nil {
		t.Fatalf("valid dimensions rejected: %v", err)
	}
	for _, dims := range []Dimensions{
		{0, 3, 3, 3, 3},
		{3, 6, 3, 3, 3},
		{3, 3, -1, 3, 3},
	} {
		if err := dims.Validate(); !errors.Is(err, ErrRatingOutOfRange) {
			t.Fatalf("Validate(%v) = %v, want ErrRatingOutOfRange", dims, err)
		}
	}
}

func TestValidateTags(t *testing.T) {
	if err := ValidateTags([]string{"分量足", "上菜快"}); err != nil {
		t.Fatalf("valid tags rejected: %v", err)
	}
	if err := ValidateTags([]string{"分量足", "分量足"}); !errors.Is(err, ErrDuplicateTag) {
		t.Fatalf("duplicate tags = %v", err)
	}
	if err := ValidateTags([]string{"不存在"}); !errors.Is(err, ErrInvalidTag) {
		t.Fatalf("invalid tag = %v", err)
	}
	if err := ValidateTags([]string{"分量足", "分量偏少", "上菜快", "等位久", "环境干净", "服务热情"}); !errors.Is(err, ErrTooManyTags) {
		t.Fatalf("too many tags = %v", err)
	}
}

func TestReviewApplyCreate(t *testing.T) {
	review := &ReviewEvent{
		StoreID:   1,
		UserID:    2,
		VisitedAt: time.Now(),
		Dimensions: Dimensions{
			Taste: 3, Value: 4, Queue: 2, Hygiene: 5, Service: 1,
		},
		Content: " 还不错 ",
		Tags:    []string{" 上菜快 ", "上菜快"},
	}
	review.ApplyCreate(DefaultStoreWeights)
	if review.Status != ReviewStatusActive {
		t.Fatalf("Status = %q", review.Status)
	}
	if review.OverallScore != 3.25 {
		t.Fatalf("OverallScore = %v, want 3.25", review.OverallScore)
	}
	if review.ScoreFormulaVersion != ScoreFormulaVersion {
		t.Fatalf("ScoreFormulaVersion = %d", review.ScoreFormulaVersion)
	}
	if review.Content != "还不错" {
		t.Fatalf("Content = %q", review.Content)
	}
	if len(review.Tags) != 1 || review.Tags[0] != "上菜快" {
		t.Fatalf("Tags = %v", review.Tags)
	}
	if err := review.Validate(); err != nil {
		t.Fatalf("Validate() = %v", err)
	}
}

func TestReviewApplyUpdateAndRevision(t *testing.T) {
	visited := time.Now().Add(-time.Hour)
	review := &ReviewEvent{
		ID:         7,
		StoreID:    1,
		UserID:     2,
		VisitedAt:  visited,
		Dimensions: Dimensions{5, 5, 5, 5, 5},
		Content:    "第一次",
		Status:     ReviewStatusActive,
	}
	review.ApplyCreate(DefaultStoreWeights)
	revision := review.MakeRevision(2, time.Now())
	if revision.ReviewEventID != 7 || revision.OverallScore != 5 || revision.Content != "第一次" {
		t.Fatalf("revision = %+v", revision)
	}

	now := time.Now()
	cmd := UpdateReview{
		VisitedAt:  now,
		Dimensions: Dimensions{1, 1, 1, 1, 1},
		Content:    "  后来变差了 ",
		Tags:       []string{"价格偏高"},
		Now:        now,
	}
	if err := review.ApplyUpdate(cmd, DefaultStoreWeights); err != nil {
		t.Fatalf("ApplyUpdate() = %v", err)
	}
	if review.OverallScore != 1 {
		t.Fatalf("OverallScore = %v, want 1", review.OverallScore)
	}
	if review.Content != "后来变差了" || !review.UpdatedAt.Equal(now) {
		t.Fatalf("review = %+v", review)
	}

	bad := UpdateReview{Dimensions: Dimensions{0, 3, 3, 3, 3}, Now: now}
	if err := review.ApplyUpdate(bad, DefaultStoreWeights); !errors.Is(err, ErrRatingOutOfRange) {
		t.Fatalf("ApplyUpdate(invalid) = %v", err)
	}
}

func TestReviewContentLimit(t *testing.T) {
	review := &ReviewEvent{
		StoreID:    1,
		UserID:     2,
		Dimensions: Dimensions{3, 3, 3, 3, 3},
		Status:     ReviewStatusActive,
		Content:    string(make([]rune, MaxReviewContentRunes+1)),
	}
	if err := review.Validate(); !errors.Is(err, ErrContentTooLong) {
		t.Fatalf("Validate() = %v", err)
	}
}
