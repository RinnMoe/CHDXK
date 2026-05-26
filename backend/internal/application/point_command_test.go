package application_test

import (
	"context"
	"errors"
	"testing"

	"jcourse/internal/application"
	"jcourse/internal/domain/account/identity"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/point"
)

func TestPointCommandService_CreateTransferSenderPaysFee(t *testing.T) {
	accountRepo := newFakePointAccountRepo()
	sender := seedFakePointAccount(accountRepo, 1, "alice@example.edu")
	seedFakePointAccount(accountRepo, 2, "bob@example.edu")
	transferRepo := &point.MockTransferRepository{}
	svc := newPointCommandService(accountRepo, transferRepo)

	got, err := svc.CreateTransfer(context.Background(), sender, application.CreatePointTransferCommand{
		RecipientEmail: "bob@example.edu",
		Amount:         100,
		FeePayer:       point.FeePayerSender,
	})
	if err != nil {
		t.Fatalf("CreateTransfer: %v", err)
	}
	if got.Fee != 2 || got.SenderDelta != -102 || got.RecipientDelta != 100 {
		t.Fatalf("transfer = %+v, want fee=2 sender=-102 recipient=100", got)
	}
	if transferRepo.SenderRecord.Reason != point.RecordReasonTransferOut || transferRepo.SenderRecord.Amount != -102 {
		t.Fatalf("sender record = %+v", transferRepo.SenderRecord)
	}
	if transferRepo.RecipientRecord.Reason != point.RecordReasonTransferIn || transferRepo.RecipientRecord.Amount != 100 {
		t.Fatalf("recipient record = %+v", transferRepo.RecipientRecord)
	}
}

func TestPointCommandService_CreateTransferRejectsInvalidCases(t *testing.T) {
	accountRepo := newFakePointAccountRepo()
	sender := seedFakePointAccount(accountRepo, 1, "alice@example.edu")
	seedFakePointAccount(accountRepo, 2, "bob@example.edu")
	svc := newPointCommandService(accountRepo, &point.MockTransferRepository{})

	tests := []struct {
		name string
		cmd  application.CreatePointTransferCommand
		want error
	}{
		{name: "amount", cmd: application.CreatePointTransferCommand{RecipientEmail: "bob@example.edu", Amount: 0, FeePayer: point.FeePayerSender}, want: application.ErrPointTransferInvalidAmount},
		{name: "fee payer", cmd: application.CreatePointTransferCommand{RecipientEmail: "bob@example.edu", Amount: 10, FeePayer: "bad"}, want: application.ErrPointTransferInvalidFeePayer},
		{name: "self", cmd: application.CreatePointTransferCommand{RecipientEmail: "alice@example.edu", Amount: 10, FeePayer: point.FeePayerSender}, want: application.ErrPointTransferSelf},
		{name: "missing recipient", cmd: application.CreatePointTransferCommand{RecipientEmail: "nobody@example.edu", Amount: 10, FeePayer: point.FeePayerSender}, want: application.ErrPointTransferRecipientNotFound},
		{name: "too small", cmd: application.CreatePointTransferCommand{RecipientEmail: "bob@example.edu", Amount: 1, FeePayer: point.FeePayerRecipient}, want: application.ErrPointTransferRecipientAmountSmall},
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

func newPointCommandService(accountRepo *identity.MockRepository, transferRepo point.TransferRepository) *application.PointCommandService {
	return application.NewPointCommandService(accountRepo, transferRepo, point.NewTransferService(point.TransferFeeConfig{RateBps: 250, MinFee: 1}), testPointUsernameDeriver())
}

func seedFakePointAccount(repo *identity.MockRepository, id int, email string) *auth.User {
	username, err := testPointUsernameDeriver().UsernameFromEmail(email)
	if err != nil {
		panic(err)
	}
	acct := &identity.Account{ID: id, Username: username, Email: email}
	repo.PutAccount("", acct)
	return &auth.User{ID: id, Role: auth.RoleUser}
}

func newFakePointAccountRepo() *identity.MockRepository {
	return identity.NewMockRepository(nil)
}

func testPointUsernameDeriver() identity.UsernameDeriver {
	return identity.NewBLAKE2bUsernameDeriver(identity.UsernameDeriverConfig{Salt: "SALT"})
}
