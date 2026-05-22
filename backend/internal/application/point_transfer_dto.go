package application

import (
	"time"

	"jcourse/internal/domain/point"
)

type PointTransferDTO struct {
	ID              int            `json:"id"`
	SenderUserID    int            `json:"sender_user_id"`
	RecipientUserID int            `json:"recipient_user_id"`
	Amount          int            `json:"amount"`
	Fee             int            `json:"fee"`
	FeePayer        point.FeePayer `json:"fee_payer"`
	SenderDelta     int            `json:"sender_delta"`
	RecipientDelta  int            `json:"recipient_delta"`
	CreatedAt       time.Time      `json:"created_at"`
}

type PointTransferPreviewDTO struct {
	Amount          int            `json:"amount"`
	Fee             int            `json:"fee"`
	FeePayer        point.FeePayer `json:"fee_payer"`
	SenderDebit     int            `json:"sender_debit"`
	RecipientCredit int            `json:"recipient_credit"`
	SenderBalance   int            `json:"sender_balance"`
	SenderRemaining int            `json:"sender_remaining"`
}

func newPointTransferDTO(t *point.Transfer) PointTransferDTO {
	return PointTransferDTO{
		ID:              t.ID,
		SenderUserID:    t.SenderUserID,
		RecipientUserID: t.RecipientUserID,
		Amount:          t.Amount,
		Fee:             t.Fee,
		FeePayer:        t.FeePayer,
		SenderDelta:     t.SenderDelta,
		RecipientDelta:  t.RecipientDelta,
		CreatedAt:       t.CreatedAt,
	}
}

func newPointTransferPreviewDTO(p *point.TransferPreview, senderBalance int) PointTransferPreviewDTO {
	return PointTransferPreviewDTO{
		Amount:          p.Amount,
		Fee:             p.Fee,
		FeePayer:        p.FeePayer,
		SenderDebit:     p.SenderDebit,
		RecipientCredit: p.RecipientCredit,
		SenderBalance:   senderBalance,
		SenderRemaining: senderBalance - p.SenderDebit,
	}
}
