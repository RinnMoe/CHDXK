//go:build test

package canteen

import (
	"bytes"
	"context"
	"io"
	"time"
)

// MockCatalog implements HallQuery and StoreQuery for application tests.
type MockCatalog struct {
	Halls        []HallView
	Stores       []StoreListItem
	StoreTotal   int64
	Detail       *StoreDetail
	Rankings     []StoreListItem
	RankingTotal int64
	Home         *HomeView

	OnListHalls      func(context.Context) ([]HallView, error)
	OnListStores     func(context.Context, StoreListFilter) ([]StoreListItem, int64, error)
	OnGetStoreDetail func(context.Context, int, int) (*StoreDetail, error)
	OnListRankings   func(context.Context, RankingFilter) ([]StoreListItem, int64, error)
	OnGetHome        func(context.Context) (*HomeView, error)
}

func (m *MockCatalog) ListHalls(ctx context.Context) ([]HallView, error) {
	if m.OnListHalls != nil {
		return m.OnListHalls(ctx)
	}
	return m.Halls, nil
}

func (m *MockCatalog) ListStores(ctx context.Context, filter StoreListFilter) ([]StoreListItem, int64, error) {
	if m.OnListStores != nil {
		return m.OnListStores(ctx, filter)
	}
	return m.Stores, m.StoreTotal, nil
}

func (m *MockCatalog) GetStoreDetail(ctx context.Context, id int, viewerID int) (*StoreDetail, error) {
	if m.OnGetStoreDetail != nil {
		return m.OnGetStoreDetail(ctx, id, viewerID)
	}
	return m.Detail, nil
}

func (m *MockCatalog) ListRankings(ctx context.Context, filter RankingFilter) ([]StoreListItem, int64, error) {
	if m.OnListRankings != nil {
		return m.OnListRankings(ctx, filter)
	}
	return m.Rankings, m.RankingTotal, nil
}

func (m *MockCatalog) GetHome(ctx context.Context) (*HomeView, error) {
	if m.OnGetHome != nil {
		return m.OnGetHome(ctx)
	}
	return m.Home, nil
}

// MockHallRepository is an in-memory HallRepository.
type MockHallRepository struct {
	NextID int
	Halls  map[int]*Hall

	OnCreate func(context.Context, *Hall) error
	OnUpdate func(context.Context, *Hall) error
	OnGet    func(context.Context, int) (*Hall, error)
}

func NewMockHallRepository() *MockHallRepository {
	return &MockHallRepository{NextID: 1, Halls: map[int]*Hall{}}
}

func (m *MockHallRepository) Create(ctx context.Context, hall *Hall) error {
	if m.OnCreate != nil {
		return m.OnCreate(ctx, hall)
	}
	if m.Halls == nil {
		m.Halls = map[int]*Hall{}
	}
	if hall.ID == 0 {
		hall.ID = m.NextID
		m.NextID++
	}
	copied := *hall
	m.Halls[hall.ID] = &copied
	return nil
}

func (m *MockHallRepository) Update(ctx context.Context, hall *Hall) error {
	if m.OnUpdate != nil {
		return m.OnUpdate(ctx, hall)
	}
	if m.Halls == nil {
		m.Halls = map[int]*Hall{}
	}
	copied := *hall
	m.Halls[hall.ID] = &copied
	return nil
}

func (m *MockHallRepository) Get(ctx context.Context, id int) (*Hall, error) {
	if m.OnGet != nil {
		return m.OnGet(ctx, id)
	}
	hall, ok := m.Halls[id]
	if !ok {
		return nil, nil
	}
	copied := *hall
	return &copied, nil
}

// MockStoreRepository is an in-memory StoreRepository.
type MockStoreRepository struct {
	NextID int
	Stores map[int]*Store

	OnCreate       func(context.Context, *Store) error
	OnUpdate       func(context.Context, *Store) error
	OnGet          func(context.Context, int) (*Store, error)
	OnUpdateStatus func(context.Context, int, StoreStatus, time.Time) error
	OnMerge        func(context.Context, int, int, time.Time) error
	OnUpdateStats  func(context.Context, int, StoreStats, time.Time) error
}

func NewMockStoreRepository() *MockStoreRepository {
	return &MockStoreRepository{NextID: 1, Stores: map[int]*Store{}}
}

func (m *MockStoreRepository) Create(ctx context.Context, store *Store) error {
	if m.OnCreate != nil {
		return m.OnCreate(ctx, store)
	}
	if m.Stores == nil {
		m.Stores = map[int]*Store{}
	}
	if store.ID == 0 {
		store.ID = m.NextID
		m.NextID++
	}
	copied := *store
	m.Stores[store.ID] = &copied
	return nil
}

func (m *MockStoreRepository) Update(ctx context.Context, store *Store) error {
	if m.OnUpdate != nil {
		return m.OnUpdate(ctx, store)
	}
	if m.Stores == nil {
		m.Stores = map[int]*Store{}
	}
	copied := *store
	m.Stores[store.ID] = &copied
	return nil
}

func (m *MockStoreRepository) Get(ctx context.Context, id int) (*Store, error) {
	if m.OnGet != nil {
		return m.OnGet(ctx, id)
	}
	store, ok := m.Stores[id]
	if !ok {
		return nil, nil
	}
	copied := *store
	return &copied, nil
}

func (m *MockStoreRepository) GetByHallAndNormalizedName(ctx context.Context, hallID int, normalizedName string) (*Store, error) {
	for _, store := range m.Stores {
		if store.HallID == hallID && store.NormalizedName == normalizedName {
			copied := *store
			return &copied, nil
		}
	}
	return nil, nil
}

func (m *MockStoreRepository) UpdateStatus(ctx context.Context, id int, status StoreStatus, now time.Time) error {
	if m.OnUpdateStatus != nil {
		return m.OnUpdateStatus(ctx, id, status, now)
	}
	if store, ok := m.Stores[id]; ok {
		store.Status = status
		store.UpdatedAt = now
	}
	return nil
}

func (m *MockStoreRepository) Merge(ctx context.Context, id, targetID int, now time.Time) error {
	if m.OnMerge != nil {
		return m.OnMerge(ctx, id, targetID, now)
	}
	if store, ok := m.Stores[id]; ok {
		store.Status = StoreStatusMerged
		store.MergedIntoStoreID = &targetID
		store.UpdatedAt = now
	}
	return nil
}

func (m *MockStoreRepository) UpdateStats(ctx context.Context, id int, stats StoreStats, now time.Time) error {
	if m.OnUpdateStats != nil {
		return m.OnUpdateStats(ctx, id, stats, now)
	}
	if store, ok := m.Stores[id]; ok {
		store.Stats = stats
		store.UpdatedAt = now
	}
	return nil
}

// MockDishRepository is an in-memory DishRepository.
type MockDishRepository struct {
	NextID int
	Dishes map[int]*Dish

	OnCreate func(context.Context, *Dish) error
}

func NewMockDishRepository() *MockDishRepository {
	return &MockDishRepository{NextID: 1, Dishes: map[int]*Dish{}}
}

func (m *MockDishRepository) Create(ctx context.Context, dish *Dish) error {
	if m.OnCreate != nil {
		return m.OnCreate(ctx, dish)
	}
	if m.Dishes == nil {
		m.Dishes = map[int]*Dish{}
	}
	if dish.ID == 0 {
		dish.ID = m.NextID
		m.NextID++
	}
	copied := *dish
	m.Dishes[dish.ID] = &copied
	return nil
}

func (m *MockDishRepository) Update(context.Context, *Dish) error { return nil }

func (m *MockDishRepository) Get(_ context.Context, id int) (*Dish, error) {
	dish, ok := m.Dishes[id]
	if !ok {
		return nil, nil
	}
	copied := *dish
	return &copied, nil
}

func (m *MockDishRepository) GetByStoreAndNormalizedName(_ context.Context, storeID int, normalizedName string) (*Dish, error) {
	for _, dish := range m.Dishes {
		if dish.StoreID == storeID && dish.NormalizedName == normalizedName {
			copied := *dish
			return &copied, nil
		}
	}
	return nil, nil
}

func (m *MockDishRepository) UpdateStatus(context.Context, int, DishStatus, time.Time) error {
	return nil
}

func (m *MockDishRepository) Merge(context.Context, int, int, time.Time) error { return nil }

func (m *MockDishRepository) UpdateStats(context.Context, int, DishStats, time.Time) error {
	return nil
}

// MockReviewRepository is an in-memory ReviewRepository.
type MockReviewRepository struct {
	NextID  int
	Reviews map[int]*ReviewEvent

	OnCreate func(context.Context, *ReviewEvent) error
}

func NewMockReviewRepository() *MockReviewRepository {
	return &MockReviewRepository{NextID: 1, Reviews: map[int]*ReviewEvent{}}
}

func (m *MockReviewRepository) Create(ctx context.Context, review *ReviewEvent) error {
	if m.OnCreate != nil {
		return m.OnCreate(ctx, review)
	}
	if m.Reviews == nil {
		m.Reviews = map[int]*ReviewEvent{}
	}
	if review.ID == 0 {
		review.ID = m.NextID
		m.NextID++
	}
	copied := *review
	m.Reviews[review.ID] = &copied
	return nil
}

func (m *MockReviewRepository) Update(_ context.Context, review *ReviewEvent, _ Revision, _ time.Time) error {
	if m.Reviews == nil {
		m.Reviews = map[int]*ReviewEvent{}
	}
	copied := *review
	m.Reviews[review.ID] = &copied
	return nil
}

func (m *MockReviewRepository) Get(_ context.Context, id int) (*ReviewEvent, error) {
	review, ok := m.Reviews[id]
	if !ok {
		return nil, nil
	}
	copied := *review
	return &copied, nil
}

func (m *MockReviewRepository) FindByClientRequestID(_ context.Context, userID int, requestID string) (*ReviewEvent, error) {
	for _, review := range m.Reviews {
		if review.UserID == userID && review.ClientRequestID == requestID {
			copied := *review
			return &copied, nil
		}
	}
	return nil, nil
}

func (m *MockReviewRepository) SoftDelete(_ context.Context, id int, now time.Time) error {
	if review, ok := m.Reviews[id]; ok {
		review.Status = ReviewStatusDeleted
		review.UpdatedAt = now
	}
	return nil
}

func (m *MockReviewRepository) Hide(_ context.Context, id int, byUserID int, reason string, now time.Time) error {
	if review, ok := m.Reviews[id]; ok {
		review.Status = ReviewStatusHidden
		review.HiddenReason = reason
		review.HiddenBy = &byUserID
		review.HiddenAt = &now
		review.UpdatedAt = now
	}
	return nil
}

func (m *MockReviewRepository) RefreshStoreStats(context.Context, int) error { return nil }

func (m *MockReviewRepository) RefreshAllStoreStats(context.Context) error { return nil }

// MockReviewQuery is an in-memory ReviewQuery.
type MockReviewQuery struct {
	Reviews  []ReviewView
	Total    int64
	Summary  *UserRatingSummary
	Count    int
	Contents []string

	OnListReviews func(context.Context, ReviewListFilter) ([]ReviewView, int64, error)
	OnGetReview   func(context.Context, int) (*ReviewView, error)
}

func (m *MockReviewQuery) ListReviews(ctx context.Context, filter ReviewListFilter) ([]ReviewView, int64, error) {
	if m.OnListReviews != nil {
		return m.OnListReviews(ctx, filter)
	}
	return m.Reviews, m.Total, nil
}

func (m *MockReviewQuery) GetReview(ctx context.Context, id int) (*ReviewView, error) {
	if m.OnGetReview != nil {
		return m.OnGetReview(ctx, id)
	}
	for i := range m.Reviews {
		if m.Reviews[i].ID == id {
			view := m.Reviews[i]
			return &view, nil
		}
	}
	return nil, nil
}

func (m *MockReviewQuery) ListUserReviews(context.Context, int, int, int) ([]ReviewView, int64, error) {
	return m.Reviews, m.Total, nil
}

func (m *MockReviewQuery) GetUserSummary(context.Context, int, int) (*UserRatingSummary, error) {
	return m.Summary, nil
}

func (m *MockReviewQuery) CountSince(context.Context, int, time.Time) (int, error) {
	return m.Count, nil
}

func (m *MockReviewQuery) ListRecentContent(context.Context, int, time.Time, int) ([]string, error) {
	return m.Contents, nil
}

// MockVoteRepository is an in-memory VoteRepository.
type MockVoteRepository struct {
	Votes map[[2]int]*Vote
}

func NewMockVoteRepository() *MockVoteRepository {
	return &MockVoteRepository{Votes: map[[2]int]*Vote{}}
}

func (m *MockVoteRepository) Get(_ context.Context, reviewEventID, userID int) (*Vote, error) {
	vote, ok := m.Votes[[2]int{reviewEventID, userID}]
	if !ok {
		return nil, nil
	}
	copied := *vote
	return &copied, nil
}

func (m *MockVoteRepository) Save(_ context.Context, vote *Vote) error {
	if m.Votes == nil {
		m.Votes = map[[2]int]*Vote{}
	}
	copied := *vote
	m.Votes[[2]int{vote.ReviewEventID, vote.UserID}] = &copied
	return nil
}

func (m *MockVoteRepository) Delete(_ context.Context, reviewEventID, userID int) error {
	delete(m.Votes, [2]int{reviewEventID, userID})
	return nil
}

// MockDishQuery is an in-memory DishQuery.
type MockDishQuery struct {
	Dishes  []DishView
	Dish    *Dish
	Reviews []DishReviewView
	Total   int64
}

func (m *MockDishQuery) ListDishes(context.Context, DishListFilter) ([]DishView, error) {
	return m.Dishes, nil
}

func (m *MockDishQuery) GetDish(context.Context, int) (*Dish, error) {
	return m.Dish, nil
}

func (m *MockDishQuery) ListDishPhotos(context.Context, int) ([]MediaView, error) {
	return nil, nil
}

func (m *MockDishQuery) ListDishReviews(context.Context, int, int, int) ([]DishReviewView, int64, error) {
	return m.Reviews, m.Total, nil
}

func (m *MockDishQuery) ListDishReviewsByUser(context.Context, int, int, int) ([]DishReviewView, int64, error) {
	return m.Reviews, m.Total, nil
}

// MockDishReviewRepository is an in-memory DishReviewRepository.
type MockDishReviewRepository struct {
	NextID  int
	Reviews []*DishReviewEvent
}

func NewMockDishReviewRepository() *MockDishReviewRepository {
	return &MockDishReviewRepository{NextID: 1}
}

func (m *MockDishReviewRepository) Create(_ context.Context, review *DishReviewEvent) error {
	if review.ID == 0 {
		review.ID = m.NextID
		m.NextID++
	}
	copied := *review
	m.Reviews = append(m.Reviews, &copied)
	return nil
}

// MockMediaRepository is an in-memory MediaRepository and MediaQuery.
type MockMediaRepository struct {
	NextID int
	Items  map[int]*Media
	Links  map[int][]int
	Public map[int]bool
}

func NewMockMediaRepository() *MockMediaRepository {
	return &MockMediaRepository{NextID: 1, Items: map[int]*Media{}, Links: map[int][]int{}, Public: map[int]bool{}}
}

func (m *MockMediaRepository) Create(_ context.Context, media *Media) error {
	if media.ID == 0 {
		media.ID = m.NextID
		m.NextID++
	}
	copied := *media
	m.Items[media.ID] = &copied
	return nil
}

func (m *MockMediaRepository) Update(_ context.Context, media *Media) error {
	copied := *media
	m.Items[media.ID] = &copied
	return nil
}

func (m *MockMediaRepository) ListByReview(_ context.Context, reviewID int) ([]Media, error) {
	result := make([]Media, 0)
	for _, id := range m.Links[reviewID] {
		if item, ok := m.Items[id]; ok {
			result = append(result, *item)
		}
	}
	return result, nil
}

func (m *MockMediaRepository) AttachToReview(_ context.Context, reviewID, mediaID int, _ int) error {
	m.Links[reviewID] = append(m.Links[reviewID], mediaID)
	return nil
}

func (m *MockMediaRepository) AttachToDish(context.Context, int, int, int, int, time.Time) error {
	return nil
}

func (m *MockMediaRepository) CountByOwnerSince(context.Context, int, time.Time) (int, error) {
	return 0, nil
}

func (m *MockMediaRepository) GetMedia(_ context.Context, id int) (*Media, error) {
	item, ok := m.Items[id]
	if !ok {
		return nil, nil
	}
	copied := *item
	return &copied, nil
}

func (m *MockMediaRepository) ListMediaByOwner(context.Context, int, int, int) ([]Media, int64, error) {
	result := make([]Media, 0, len(m.Items))
	for _, item := range m.Items {
		result = append(result, *item)
	}
	return result, int64(len(result)), nil
}

func (m *MockMediaRepository) ListMediaByStatus(context.Context, MediaFilter) ([]Media, int64, error) {
	result := make([]Media, 0, len(m.Items))
	for _, item := range m.Items {
		result = append(result, *item)
	}
	return result, int64(len(result)), nil
}

func (m *MockMediaRepository) IsMediaPublic(_ context.Context, id int) (bool, error) {
	return m.Public[id], nil
}

// MockReportRepository is an in-memory ReportRepository.
type MockReportRepository struct {
	NextID  int
	Reports map[int]*Report
}

func NewMockReportRepository() *MockReportRepository {
	return &MockReportRepository{NextID: 1, Reports: map[int]*Report{}}
}

func (m *MockReportRepository) Create(_ context.Context, report *Report) error {
	if report.ID == 0 {
		report.ID = m.NextID
		m.NextID++
	}
	copied := *report
	m.Reports[report.ID] = &copied
	return nil
}

func (m *MockReportRepository) Get(_ context.Context, id int) (*Report, error) {
	report, ok := m.Reports[id]
	if !ok {
		return nil, nil
	}
	copied := *report
	return &copied, nil
}

func (m *MockReportRepository) List(context.Context, ReportFilter) ([]Report, int64, error) {
	result := make([]Report, 0, len(m.Reports))
	for _, report := range m.Reports {
		result = append(result, *report)
	}
	return result, int64(len(result)), nil
}

func (m *MockReportRepository) Resolve(_ context.Context, id int, handlerID int, status ReportStatus, note string, now time.Time) error {
	report, ok := m.Reports[id]
	if !ok {
		return ErrReportNotFound
	}
	report.Status = status
	report.HandledBy = &handlerID
	report.HandledAt = &now
	report.ResolutionNote = note
	return nil
}

// MockMediaStorage is an in-memory MediaStorage.
type MockMediaStorage struct {
	Files map[string][]byte
}

func NewMockMediaStorage() *MockMediaStorage {
	return &MockMediaStorage{Files: map[string][]byte{}}
}

func (m *MockMediaStorage) Put(_ context.Context, key string, reader io.Reader) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	m.Files[key] = data
	return nil
}

func (m *MockMediaStorage) Open(_ context.Context, key string) (io.ReadCloser, error) {
	data, ok := m.Files[key]
	if !ok {
		return nil, ErrMediaNotFound
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (m *MockMediaStorage) Delete(_ context.Context, key string) error {
	delete(m.Files, key)
	return nil
}

// MockImageProcessor returns a fixed processed image.
type MockImageProcessor struct {
	Result ProcessedImage
	Err    error
}

func (m MockImageProcessor) Process([]byte, int) (ProcessedImage, error) {
	return m.Result, m.Err
}
