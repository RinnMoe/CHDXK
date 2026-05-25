//go:build test

package point

import "context"

type MockTransferRepository struct {
	Transfer        *Transfer
	SenderRecord    Record
	RecipientRecord Record

	OnCreateTransfer func(context.Context, *Transfer, Record, Record) error
}

func (r *MockTransferRepository) CreateTransfer(ctx context.Context, t *Transfer, senderRecord Record, recipientRecord Record) error {
	if r.OnCreateTransfer != nil {
		return r.OnCreateTransfer(ctx, t, senderRecord, recipientRecord)
	}
	copy := *t
	if copy.ID == 0 {
		copy.ID = 99
		t.ID = copy.ID
	}
	r.Transfer = &copy
	r.SenderRecord = senderRecord
	r.RecipientRecord = recipientRecord
	return nil
}
