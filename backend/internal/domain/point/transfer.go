package point

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	FeePayerSender    FeePayer = "sender"
	FeePayerRecipient FeePayer = "recipient"

	RecordReasonTransferOut RecordReason = "transfer_out"
	RecordReasonTransferIn  RecordReason = "transfer_in"
)

type FeePayer string

var (
	ErrInsufficientBalance          = errors.New("insufficient point balance")
	ErrTransferInvalidAmount        = errors.New("point transfer amount must be positive")
	ErrTransferInvalidFeePayer      = errors.New("invalid point transfer fee payer")
	ErrTransferRecipientAmountSmall = errors.New("point transfer amount is too small after fee")
)

type TransferFeeConfig struct {
	RateBps int
	MinFee  int
}

type UserRef struct {
	ID       int
	Username string
}

type TransferPreview struct {
	Amount          int
	Fee             int
	FeePayer        FeePayer
	SenderDebit     int
	RecipientCredit int
}

type TransferService struct {
	feeConfig TransferFeeConfig
}

func NewTransferService(feeConfig TransferFeeConfig) *TransferService {
	return &TransferService{feeConfig: normalizeFeeConfig(feeConfig)}
}

type Transfer struct {
	ID              int
	SenderUserID    int
	RecipientUserID int
	Amount          int
	Fee             int
	FeePayer        FeePayer
	SenderDelta     int
	RecipientDelta  int
	CreatedAt       time.Time
}

func (s *TransferService) Preview(amount int, feePayer FeePayer) (*TransferPreview, error) {
	if amount <= 0 {
		return nil, ErrTransferInvalidAmount
	}
	feePayer = normalizeFeePayer(feePayer)
	if feePayer != FeePayerSender && feePayer != FeePayerRecipient {
		return nil, ErrTransferInvalidFeePayer
	}

	fee, senderDebit, recipientCredit := calculateTransferAmounts(amount, feePayer, s.feeConfig)
	if recipientCredit <= 0 {
		return nil, ErrTransferRecipientAmountSmall
	}

	return &TransferPreview{
		Amount:          amount,
		Fee:             fee,
		FeePayer:        feePayer,
		SenderDebit:     senderDebit,
		RecipientCredit: recipientCredit,
	}, nil
}

func (s *TransferService) NewTransfer(sender UserRef, recipient UserRef, amount int, feePayer FeePayer, now time.Time) (*Transfer, Record, Record, error) {
	preview, err := s.Preview(amount, feePayer)
	if err != nil {
		return nil, Record{}, Record{}, err
	}

	transfer := &Transfer{
		SenderUserID:    sender.ID,
		RecipientUserID: recipient.ID,
		Amount:          preview.Amount,
		Fee:             preview.Fee,
		FeePayer:        preview.FeePayer,
		SenderDelta:     -preview.SenderDebit,
		RecipientDelta:  preview.RecipientCredit,
		CreatedAt:       now,
	}

	senderRecord := Record{
		UserID:      sender.ID,
		Reason:      RecordReasonTransferOut,
		Amount:      transfer.SenderDelta,
		Description: fmt.Sprintf("转账给 %s", recipient.Username),
		CreatedAt:   now,
	}
	recipientRecord := Record{
		UserID:      recipient.ID,
		Reason:      RecordReasonTransferIn,
		Amount:      transfer.RecipientDelta,
		Description: fmt.Sprintf("收到 %s 的转账", sender.Username),
		CreatedAt:   now,
	}

	return transfer, senderRecord, recipientRecord, nil
}

func normalizeFeePayer(feePayer FeePayer) FeePayer {
	feePayer = FeePayer(strings.TrimSpace(string(feePayer)))
	if feePayer == FeePayer("") {
		return FeePayerSender
	}
	return feePayer
}

func normalizeFeeConfig(c TransferFeeConfig) TransferFeeConfig {
	if c.RateBps < 0 {
		c.RateBps = 0
	}
	if c.MinFee < 0 {
		c.MinFee = 0
	}
	return c
}

func calculateTransferAmounts(amount int, feePayer FeePayer, feeConfig TransferFeeConfig) (fee int, senderDebit int, recipientCredit int) {
	if feePayer == FeePayerRecipient {
		fee = recipientPaidFee(amount, feeConfig)
		return fee, amount, amount - fee
	}
	fee = feeForReceived(amount, feeConfig)
	return fee, amount + fee, amount
}

func feeForReceived(received int, feeConfig TransferFeeConfig) int {
	fee := max(received*feeConfig.RateBps/10000, feeConfig.MinFee)
	return fee
}

func recipientPaidFee(amount int, feeConfig TransferFeeConfig) int {
	low, high := 0, amount
	for low < high {
		mid := (low + high + 1) / 2
		if mid+feeForReceived(mid, feeConfig) <= amount {
			low = mid
		} else {
			high = mid - 1
		}
	}
	return amount - low
}

type TransferRepository interface {
	CreateTransfer(ctx context.Context, t *Transfer, senderRecord Record, recipientRecord Record) error
}
