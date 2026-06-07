package application

import (
	"context"

	"jcourse/internal/domain/account/identity"
	"jcourse/internal/domain/point"
	"jcourse/pkg/apperr"
)

type PointRecordListFilter struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

type PointQueryService struct {
	query       point.Query
	accountRepo identity.Repository
	usernames   identity.UsernameDeriver
}

func NewPointQueryService(query point.Query, accountRepo identity.Repository, usernames identity.UsernameDeriver) *PointQueryService {
	return &PointQueryService{query: query, accountRepo: accountRepo, usernames: usernames}
}

func (s *PointQueryService) GetUserPoints(ctx context.Context, userID int, f PointRecordListFilter) (*PointSummaryDTO, error) {
	total, err := s.query.SumByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	records, recordTotal, err := s.query.FindRecordsByUser(ctx, point.RecordFilter{
		UserID:   userID,
		Page:     f.Page,
		PageSize: f.PageSize,
	})
	if err != nil {
		return nil, err
	}

	items := make([]PointRecordDTO, len(records))
	for i, r := range records {
		items[i] = newPointRecordDTO(&r)
	}

	return &PointSummaryDTO{
		Total: total,
		Records: PaginatedResult[PointRecordDTO]{
			Items:    items,
			Total:    recordTotal,
			Page:     f.Page,
			PageSize: f.PageSize,
		},
	}, nil
}

var ErrPointUserNotFound = apperr.ErrPointUserNotFound

func (s *PointQueryService) GetUserPointsByEmail(ctx context.Context, email string) (int, error) {
	username, err := s.usernames.UsernameFromEmail(email)
	if err != nil {
		return 0, err
	}
	u, err := s.accountRepo.FindByUsername(ctx, username)
	if err != nil {
		return 0, err
	}
	if u == nil {
		return 0, ErrPointUserNotFound
	}
	return s.query.SumByUser(ctx, u.ID)
}
