//go:build test

package review

import (
	"context"

	"jcourse/internal/domain/point"
)

type MockReviewRepository struct {
	NextID  int
	Reviews map[int]*Review

	OnCreate                func(context.Context, *Review) error
	OnCreateWithReward      func(context.Context, *Review, []point.Reward) (CreateResult, error)
	OnUpdate                func(context.Context, *Review, Revision) error
	OnUpdateModeratorRemark func(context.Context, int, string) error
	OnDelete                func(context.Context, *Review) error
	OnGet                   func(context.Context, int) (*Review, error)
}

func NewMockReviewRepository() *MockReviewRepository {
	return &MockReviewRepository{NextID: 1, Reviews: map[int]*Review{}}
}

func (r *MockReviewRepository) Create(ctx context.Context, rv *Review) error {
	if r.OnCreate != nil {
		return r.OnCreate(ctx, rv)
	}
	r.ensureMap()
	copy := *rv
	if copy.ID == 0 {
		copy.ID = r.NextID
		r.NextID++
	}
	r.Reviews[copy.ID] = &copy
	rv.ID = copy.ID
	return nil
}

func (r *MockReviewRepository) CreateWithReward(ctx context.Context, rv *Review, rewards []point.Reward) (CreateResult, error) {
	if r.OnCreateWithReward != nil {
		return r.OnCreateWithReward(ctx, rv, rewards)
	}
	if err := r.Create(ctx, rv); err != nil {
		return CreateResult{}, err
	}
	if len(rewards) == 0 {
		return CreateResult{}, nil
	}
	result := CreateResult{RewardIDs: make([]int, 0, len(rewards))}
	for i := range rewards {
		if rewards[i].ID == 0 {
			rewards[i].ID = 1000 + rv.ID + i
		}
		result.RewardIDs = append(result.RewardIDs, rewards[i].ID)
	}
	return result, nil
}

func (r *MockReviewRepository) Update(ctx context.Context, rv *Review, revision Revision) error {
	if r.OnUpdate != nil {
		return r.OnUpdate(ctx, rv, revision)
	}
	r.ensureMap()
	copy := *rv
	r.Reviews[copy.ID] = &copy
	return nil
}

func (r *MockReviewRepository) UpdateModeratorRemark(ctx context.Context, reviewID int, moderatorRemark string) error {
	if r.OnUpdateModeratorRemark != nil {
		return r.OnUpdateModeratorRemark(ctx, reviewID, moderatorRemark)
	}
	r.ensureMap()
	if rv, ok := r.Reviews[reviewID]; ok {
		copy := *rv
		copy.ModeratorRemark = moderatorRemark
		r.Reviews[reviewID] = &copy
	}
	return nil
}

func (r *MockReviewRepository) Delete(ctx context.Context, rv *Review) error {
	if r.OnDelete != nil {
		return r.OnDelete(ctx, rv)
	}
	r.ensureMap()
	delete(r.Reviews, rv.ID)
	return nil
}

func (r *MockReviewRepository) Get(ctx context.Context, reviewID int) (*Review, error) {
	if r.OnGet != nil {
		return r.OnGet(ctx, reviewID)
	}
	r.ensureMap()
	rv, ok := r.Reviews[reviewID]
	if !ok {
		return nil, nil
	}
	copy := *rv
	return &copy, nil
}

func (r *MockReviewRepository) ensureMap() {
	if r.NextID == 0 {
		r.NextID = 1
	}
	if r.Reviews == nil {
		r.Reviews = map[int]*Review{}
	}
}

type MockVoteRepository struct {
	Existing   *Vote
	Votes      map[int]Vote
	TodayCount int64
	BatchCalls int

	OnFindByReviewAndUser  func(context.Context, int, int) (*Vote, error)
	OnFindByReviewsAndUser func(context.Context, []int, int) (map[int]Vote, error)
	OnCountTodayByUser     func(context.Context, int) (int64, error)
	OnSave                 func(context.Context, *Vote) error
	OnDelete               func(context.Context, int, int) error
}

func (r *MockVoteRepository) FindByReviewAndUser(ctx context.Context, reviewID, userID int) (*Vote, error) {
	if r.OnFindByReviewAndUser != nil {
		return r.OnFindByReviewAndUser(ctx, reviewID, userID)
	}
	if r.Existing != nil && r.Existing.ReviewID == reviewID && r.Existing.UserID == userID {
		copy := *r.Existing
		return &copy, nil
	}
	if vote, ok := r.Votes[reviewID]; ok && vote.UserID == userID {
		copy := vote
		return &copy, nil
	}
	return nil, nil
}

func (r *MockVoteRepository) FindByReviewsAndUser(ctx context.Context, reviewIDs []int, userID int) (map[int]Vote, error) {
	if r.OnFindByReviewsAndUser != nil {
		return r.OnFindByReviewsAndUser(ctx, reviewIDs, userID)
	}
	r.BatchCalls++
	result := make(map[int]Vote)
	for _, reviewID := range reviewIDs {
		vote, ok := r.Votes[reviewID]
		if ok && vote.UserID == userID {
			result[reviewID] = vote
		}
	}
	return result, nil
}

func (r *MockVoteRepository) CountTodayByUser(ctx context.Context, userID int) (int64, error) {
	if r.OnCountTodayByUser != nil {
		return r.OnCountTodayByUser(ctx, userID)
	}
	return r.TodayCount, nil
}

func (r *MockVoteRepository) Save(ctx context.Context, vote *Vote) error {
	if r.OnSave != nil {
		return r.OnSave(ctx, vote)
	}
	copy := *vote
	r.Existing = &copy
	if r.Votes == nil {
		r.Votes = map[int]Vote{}
	}
	r.Votes[vote.ReviewID] = copy
	return nil
}

func (r *MockVoteRepository) Delete(ctx context.Context, reviewID, userID int) error {
	if r.OnDelete != nil {
		return r.OnDelete(ctx, reviewID, userID)
	}
	if r.Existing != nil && r.Existing.ReviewID == reviewID && r.Existing.UserID == userID {
		r.Existing = nil
	}
	delete(r.Votes, reviewID)
	return nil
}
