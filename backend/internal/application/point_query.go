package application

import (
	"context"
	"errors"

	"jcourse/internal/domain/account"
	"jcourse/internal/domain/point"
)

type PointRecordListFilter struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

type PointQueryService struct {
	query           point.Query
	accountRepo     account.AccountRepository
	transferService *point.TransferService
}

func NewPointQueryService(query point.Query, accountRepo account.AccountRepository, transferService *point.TransferService) *PointQueryService {
	return &PointQueryService{query: query, accountRepo: accountRepo, transferService: transferService}
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

type PreviewTransferParams struct {
	Amount   int            `json:"amount"`
	FeePayer point.FeePayer `json:"fee_payer"`
}

func (s *PointQueryService) PreviewTransfer(ctx context.Context, userID int, params PreviewTransferParams) (*PointTransferPreviewDTO, error) {
	senderBalance, err := s.query.SumByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	preview, err := s.transferService.Preview(params.Amount, params.FeePayer)
	if err != nil {
		return nil, err
	}
	dto := newPointTransferPreviewDTO(preview, senderBalance)
	return &dto, nil
}

var ErrPointUserNotFound = errors.New("user not found")

func (s *PointQueryService) GetUserPointsByEmail(ctx context.Context, email string) (int, error) {
	u, err := s.accountRepo.FindByEmail(ctx, email)
	if err != nil {
		return 0, err
	}
	if u == nil {
		return 0, ErrPointUserNotFound
	}
	return s.query.SumByUser(ctx, u.ID)
}
