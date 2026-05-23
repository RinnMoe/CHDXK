package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"jcourse/internal/application"
	"jcourse/internal/domain/account"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/point"
)

func TestPointCommandService_CreateTransferSenderPaysFee(t *testing.T) {
	accountRepo := newFakePointAccountRepo()
	sender := seedFakePointAccount(accountRepo, 1, "alice@example.edu")
	seedFakePointAccount(accountRepo, 2, "bob@example.edu")
	transferRepo := &fakePointTransferRepo{}
	svc := newPointCommandService(accountRepo, transferRepo)

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
	accountRepo := newFakePointAccountRepo()
	sender := seedFakePointAccount(accountRepo, 1, "alice@example.edu")
	seedFakePointAccount(accountRepo, 2, "bob@example.edu")
	svc := newPointCommandService(accountRepo, &fakePointTransferRepo{})

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

func newPointCommandService(accountRepo *fakePointAccountRepo, transferRepo point.TransferRepository) *application.PointCommandService {
	return application.NewPointCommandService(accountRepo, transferRepo, point.NewTransferService(point.TransferFeeConfig{RateBps: 250, MinFee: 1}))
}

func seedFakePointAccount(repo *fakePointAccountRepo, id int, email string) *auth.User {
	acct := &account.Account{ID: id, Username: email, Email: email}
	repo.accountsByID[id] = acct
	repo.accountsByUsername[email] = acct
	return &auth.User{ID: id, Role: auth.RoleUser}
}

type fakePointAccountRepo struct {
	accountsByID       map[int]*account.Account
	accountsByUsername map[string]*account.Account
}

func newFakePointAccountRepo() *fakePointAccountRepo {
	return &fakePointAccountRepo{accountsByID: map[int]*account.Account{}, accountsByUsername: map[string]*account.Account{}}
}

func (r *fakePointAccountRepo) Create(_ context.Context, _ *account.Account) error { return nil }
func (r *fakePointAccountRepo) Update(_ context.Context, _ *account.Account) error { return nil }
func (r *fakePointAccountRepo) TouchLastSeen(_ context.Context, _ int, _ time.Time) error {
	return nil
}
func (r *fakePointAccountRepo) FindByID(_ context.Context, id int) (*account.Account, error) {
	return r.accountsByID[id], nil
}
func (r *fakePointAccountRepo) FindByUsername(_ context.Context, username string) (*account.Account, error) {
	return r.accountsByUsername[username], nil
}
func (r *fakePointAccountRepo) FindByEmail(_ context.Context, email string) (*account.Account, error) {
	return r.accountsByUsername[email], nil
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
