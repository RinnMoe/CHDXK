package reshub

import (
	"context"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/point"
)

type PointQuery struct {
	provider   point.ExternalPointProvider
	identities auth.IdentityProjectionRepository
}

func NewPointQuery(provider point.ExternalPointProvider, identities auth.IdentityProjectionRepository) *PointQuery {
	return &PointQuery{provider: provider, identities: identities}
}

func (q *PointQuery) SumByUser(ctx context.Context, userID int) (int, error) {
	subject, err := q.identities.SubjectByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}
	return q.provider.Balance(ctx, subject)
}

func (q *PointQuery) FindRecordsByUser(ctx context.Context, filter point.RecordFilter) ([]point.Record, int64, error) {
	subject, err := q.identities.SubjectByUserID(ctx, filter.UserID)
	if err != nil {
		return nil, 0, err
	}
	operations, total, err := q.provider.ListOperations(ctx, subject, filter.Page, filter.PageSize)
	if err != nil {
		return nil, 0, err
	}
	records := make([]point.Record, len(operations))
	for i := range operations {
		records[i] = point.Record{
			ExternalID:  operations[i].ID,
			UserID:      filter.UserID,
			Reason:      point.RecordReason(operations[i].Reason),
			Amount:      operations[i].Amount,
			Description: operations[i].Description,
			CreatedAt:   operations[i].OccurredAt,
		}
		if records[i].Description == "" {
			records[i].Description = operations[i].Reason
		}
	}
	return records, total, nil
}

var _ point.Query = (*PointQuery)(nil)
