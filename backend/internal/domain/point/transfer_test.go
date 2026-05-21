package point

import (
	"errors"
	"testing"
	"time"
)

func TestTransferService_Preview(t *testing.T) {
	svc := NewTransferService(TransferFeeConfig{RateBps: 250, MinFee: 1})

	tests := []struct {
		name                string
		amount              int
		feePayer            FeePayer
		wantFee             int
		wantFeePayer        FeePayer
		wantSenderDebit     int
		wantRecipientCredit int
	}{
		{name: "sender pays", amount: 100, feePayer: FeePayerSender, wantFee: 2, wantFeePayer: FeePayerSender, wantSenderDebit: 102, wantRecipientCredit: 100},
		{name: "recipient pays", amount: 100, feePayer: FeePayerRecipient, wantFee: 2, wantFeePayer: FeePayerRecipient, wantSenderDebit: 100, wantRecipientCredit: 98},
		{name: "default fee payer", amount: 100, feePayer: FeePayer(""), wantFee: 2, wantFeePayer: FeePayerSender, wantSenderDebit: 102, wantRecipientCredit: 100},
		{name: "min fee", amount: 10, feePayer: FeePayerSender, wantFee: 1, wantFeePayer: FeePayerSender, wantSenderDebit: 11, wantRecipientCredit: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.Preview(tt.amount, tt.feePayer)
			if err != nil {
				t.Fatalf("Preview: %v", err)
			}
			if got.Fee != tt.wantFee || got.FeePayer != tt.wantFeePayer || got.SenderDebit != tt.wantSenderDebit || got.RecipientCredit != tt.wantRecipientCredit {
				t.Fatalf("preview = %+v", got)
			}
		})
	}
}

func TestTransferService_PreviewRejectsInvalidInput(t *testing.T) {
	svc := NewTransferService(TransferFeeConfig{RateBps: 250, MinFee: 1})

	tests := []struct {
		name     string
		amount   int
		feePayer FeePayer
		want     error
	}{
		{name: "amount", amount: 0, feePayer: FeePayerSender, want: ErrTransferInvalidAmount},
		{name: "fee payer", amount: 10, feePayer: FeePayer("bad"), want: ErrTransferInvalidFeePayer},
		{name: "too small", amount: 1, feePayer: FeePayerRecipient, want: ErrTransferRecipientAmountSmall},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Preview(tt.amount, tt.feePayer)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Preview error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestTransferService_NewTransfer(t *testing.T) {
	svc := NewTransferService(TransferFeeConfig{RateBps: 250, MinFee: 1})
	now := time.Now()

	transfer, senderRecord, recipientRecord, err := svc.NewTransfer(
		UserRef{ID: 1, Username: "alice@example.edu"},
		UserRef{ID: 2, Username: "bob@example.edu"},
		100,
		FeePayerSender,
		now,
	)
	if err != nil {
		t.Fatalf("NewTransfer: %v", err)
	}
	if transfer.SenderDelta != -102 || transfer.RecipientDelta != 100 || transfer.Fee != 2 {
		t.Fatalf("transfer = %+v", transfer)
	}
	if senderRecord.UserID != 1 || senderRecord.Reason != RecordReasonTransferOut || senderRecord.Amount != -102 || senderRecord.Description != "转账给 bob@example.edu" {
		t.Fatalf("sender record = %+v", senderRecord)
	}
	if recipientRecord.UserID != 2 || recipientRecord.Reason != RecordReasonTransferIn || recipientRecord.Amount != 100 || recipientRecord.Description != "收到 alice@example.edu 的转账" {
		t.Fatalf("recipient record = %+v", recipientRecord)
	}
}
