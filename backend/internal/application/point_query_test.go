package application_test

import (
	"context"
	"testing"
	"time"

	"jcourse/internal/application"
	"jcourse/internal/domain/point"
)

func TestPointQueryService_GetUserPoints(t *testing.T) {
	createdAt := time.Now()
	repo := &fakePointQuery{
		total: 15,
		records: []point.Record{
			{Reason: point.RecordReason("review_created"), Amount: 10, Description: "发布课程评价", CreatedAt: createdAt},
			{Reason: point.RecordReason("review_liked"), Amount: 5, Description: "评价收到赞同", CreatedAt: createdAt.Add(time.Second)},
		},
		recordTotal: 2,
	}
	svc := application.NewPointQueryService(repo)

	result, err := svc.GetUserPoints(context.Background(), 7, application.PointRecordListFilter{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("GetUserPoints: %v", err)
	}
	if repo.userID != 7 {
		t.Fatalf("queried user id = %d, want 7", repo.userID)
	}
	if result.Total != 15 {
		t.Fatalf("total = %d, want 15", result.Total)
	}
	if result.Records.Total != 2 || result.Records.Page != 1 || result.Records.PageSize != 20 {
		t.Fatalf("records page = %+v", result.Records)
	}
	if len(result.Records.Items) != 2 || result.Records.Items[0].Reason != "review_created" {
		t.Fatalf("items = %+v", result.Records.Items)
	}
}

type fakePointQuery struct {
	userID      int
	total       int
	records     []point.Record
	recordTotal int64
}

func (q *fakePointQuery) SumByUser(_ context.Context, userID int) (int, error) {
	q.userID = userID
	return q.total, nil
}

func (q *fakePointQuery) FindRecordsByUser(_ context.Context, filter point.RecordFilter) ([]point.Record, int64, error) {
	q.userID = filter.UserID
	return q.records, q.recordTotal, nil
}
