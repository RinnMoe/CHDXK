package application

import (
	"context"

	"jcourse/internal/domain/point"
)

type PointRecordListFilter struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

type PointQueryService struct {
	query point.Query
}

func NewPointQueryService(query point.Query) *PointQueryService {
	return &PointQueryService{query: query}
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
