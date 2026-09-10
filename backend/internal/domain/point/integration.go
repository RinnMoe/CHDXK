package point

import (
	"context"
	"time"

	"jcourse/pkg/apperr"
)

var (
	ErrNoPointAccount    = apperr.NotFound("该账户没有积分账户")
	ErrOperationNotFound = apperr.NotFound("积分操作不存在")
	ErrOperationConflict = apperr.Conflict("积分业务键冲突")
	ErrPointUnavailable  = apperr.ServiceUnavailable("积分服务暂不可用，请稍后重试")
)

// ExternalOperation is the normalized result returned by res-hub. Keeping
// the remote operation id as a string avoids imposing an integer format on
// the integration boundary.
type ExternalOperation struct {
	ID           string
	ClientID     string
	Subject      string
	Amount       int
	Reason       string
	Description  string
	OperationKey string
	BalanceAfter int
	OccurredAt   time.Time
	OriginalID   string
}

type ExternalPointProvider interface {
	Balance(ctx context.Context, subject string) (int, error)
	ListOperations(ctx context.Context, subject string, page, pageSize int) ([]ExternalOperation, int64, error)
	Change(ctx context.Context, subject string, amount int, reason, operationKey string) (ExternalOperation, error)
	FindOperation(ctx context.Context, subject, operationKey string) (*ExternalOperation, error)
	// Reverse accepts the immutable operation_id returned by res-hub. The
	// operation key is only used for idempotent lookup before a reversal.
	Reverse(ctx context.Context, subject, operationID string) (ExternalOperation, error)
}
