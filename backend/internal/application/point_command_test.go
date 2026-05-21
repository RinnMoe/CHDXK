package application_test

import (
	"context"
	"errors"
	"testing"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/point"
)

func TestPointCommandService_CreateTransferSenderPaysFee(t *testing.T) {
	userRepo := newFakeUserRepo()
	sender := seedFakePointUser(userRepo, 1, "alice@example.edu")
	seedFakePointUser(userRepo, 2, "bob@example.edu")
	transferRepo := &fakePointTransferRepo{}
	svc := newPointCommandService(userRepo, transferRepo)

	got, err := svc.CreateTransfer(context.Background(), sender, application.CreatePointTransferCommand{
		RecipientUsername: "bob@example.edu",
		Amount:            100,
		FeePayer:          point.FeePayerSender,
	})
	if err != nil {
		t.Fatalf("CreateTransfer: %v", err)
	}
	if got.Fee != 2 || got.SenderDelta != -102 || got.RecipientDelta != 100 {
		t.Fatalf("transfer = %+v, want fee=2 sender=-102 recipient=100", got)
	}
	if transferRepo.senderRecord.Reason != point.RecordReasonTransferOut || transferRepo.senderRecord.Amount != -102 {
		t.Fatalf("sender record = %+v", transferRepo.senderRecord)
	}
	if transferRepo.recipientRecord.Reason != point.RecordReasonTransferIn || transferRepo.recipientRecord.Amount != 100 {
		t.Fatalf("recipient record = %+v", transferRepo.recipientRecord)
	}
}

func TestPointCommandService_CreateTransferRejectsInvalidCases(t *testing.T) {
	userRepo := newFakeUserRepo()
	sender := seedFakePointUser(userRepo, 1, "alice@example.edu")
	seedFakePointUser(userRepo, 2, "bob@example.edu")
	svc := newPointCommandService(userRepo, &fakePointTransferRepo{})

	tests := []struct {
		name string
		cmd  application.CreatePointTransferCommand
		want error
	}{
		{name: "amount", cmd: application.CreatePointTransferCommand{RecipientUsername: "bob@example.edu", Amount: 0, FeePayer: point.FeePayerSender}, want: application.ErrPointTransferInvalidAmount},
		{name: "fee payer", cmd: application.CreatePointTransferCommand{RecipientUsername: "bob@example.edu", Amount: 10, FeePayer: "bad"}, want: application.ErrPointTransferInvalidFeePayer},
		{name: "self", cmd: application.CreatePointTransferCommand{RecipientUsername: "alice@example.edu", Amount: 10, FeePayer: point.FeePayerSender}, want: application.ErrPointTransferSelf},
		{name: "missing recipient", cmd: application.CreatePointTransferCommand{RecipientUsername: "nobody@example.edu", Amount: 10, FeePayer: point.FeePayerSender}, want: application.ErrPointTransferRecipientNotFound},
		{name: "too small", cmd: application.CreatePointTransferCommand{RecipientUsername: "bob@example.edu", Amount: 1, FeePayer: point.FeePayerRecipient}, want: application.ErrPointTransferRecipientAmountSmall},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateTransfer(context.Background(), sender, tt.cmd)
			if !errors.Is(err, tt.want) {
				t.Fatalf("CreateTransfer error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestPointCommandService_PreviewTransfer(t *testing.T) {
	svc := newPointCommandService(newFakeUserRepo(), &fakePointTransferRepo{})

	got, err := svc.PreviewTransfer(application.CreatePointTransferCommand{Amount: 100, FeePayer: point.FeePayerRecipient})
	if err != nil {
		t.Fatalf("PreviewTransfer: %v", err)
	}
	if got.Fee != 2 || got.SenderDebit != 100 || got.RecipientCredit != 98 {
		t.Fatalf("preview = %+v", got)
	}
}

func newPointCommandService(userRepo *fakeUserRepo, transferRepo point.TransferRepository) *application.PointCommandService {
	return application.NewPointCommandService(userRepo, transferRepo, application.PointTransferFeeConfig{RateBps: 250, MinFee: 1})
}

func seedFakePointUser(repo *fakeUserRepo, id int, email string) *auth.User {
	u := &auth.User{ID: id, Username: email, Email: email, Role: auth.RoleUser}
	repo.usersByID[id] = u
	repo.usersByEmail[email] = u
	return u
}

type fakePointTransferRepo struct {
	transfer        *point.Transfer
	senderRecord    point.Record
	recipientRecord point.Record
}

func (r *fakePointTransferRepo) CreateTransfer(_ context.Context, t *point.Transfer, senderRecord point.Record, recipientRecord point.Record) error {
	copy := *t
	copy.ID = 99
	t.ID = copy.ID
	r.transfer = &copy
	r.senderRecord = senderRecord
	r.recipientRecord = recipientRecord
	return nil
}
