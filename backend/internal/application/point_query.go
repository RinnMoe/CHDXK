package application

import (
	"context"
	"errors"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/point"
)

type PointRecordListFilter struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

type PointQueryService struct {
	query    point.Query
	userRepo auth.UserRepository
}

func NewPointQueryService(query point.Query, userRepo auth.UserRepository) *PointQueryService {
	return &PointQueryService{query: query, userRepo: userRepo}
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

var ErrPointUserNotFound = errors.New("user not found")

func (s *PointQueryService) GetUserPointsByEmail(ctx context.Context, email string) (int, error) {
	u, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return 0, err
	}
	if u == nil {
		return 0, ErrPointUserNotFound
	}
	return s.query.SumByUser(ctx, u.ID)
}
