package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"jcourse/internal/domain/account"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/point"
)

var (
	ErrPointTransferInvalidAmount        = point.ErrTransferInvalidAmount
	ErrPointTransferInvalidFeePayer      = point.ErrTransferInvalidFeePayer
	ErrPointTransferSelf                 = point.ErrTransferSelf
	ErrPointTransferRecipientNotFound    = errors.New("point transfer recipient not found")
	ErrPointTransferRecipientAmountSmall = point.ErrTransferRecipientAmountSmall
)

type CreatePointTransferCommand struct {
	RecipientEmail string         `json:"recipient_email"`
	Amount         int            `json:"amount"`
	FeePayer       point.FeePayer `json:"fee_payer"`
}

type PointCommandService struct {
	accountRepo     account.AccountRepository
	transferRepo    point.TransferRepository
	transferService *point.TransferService
	usernames       account.UsernameDeriver
}

func NewPointCommandService(accountRepo account.AccountRepository, transferRepo point.TransferRepository, transferService *point.TransferService, usernames account.UsernameDeriver) *PointCommandService {
	return &PointCommandService{
		accountRepo:     accountRepo,
		transferRepo:    transferRepo,
		transferService: transferService,
		usernames:       usernames,
	}
}

func (s *PointCommandService) CreateTransfer(ctx context.Context, sender *auth.User, cmd CreatePointTransferCommand) (*PointTransferDTO, error) {
	senderAccount, err := s.accountRepo.FindByID(ctx, sender.ID)
	if err != nil {
		return nil, err
	}
	if senderAccount == nil {
		return nil, ErrPointTransferRecipientNotFound
	}

	recipientUsername, err := s.usernames.UsernameFromEmail(strings.TrimSpace(cmd.RecipientEmail))
	if err != nil {
		return nil, err
	}

	recipient, err := s.accountRepo.FindByUsername(ctx, recipientUsername)
	if err != nil {
		return nil, err
	}
	if recipient == nil {
		return nil, ErrPointTransferRecipientNotFound
	}

	now := time.Now()
	transfer, senderRecord, recipientRecord, err := s.transferService.NewTransfer(
		sender.ID,
		recipient.ID,
		cmd.Amount,
		cmd.FeePayer,
		now,
	)
	if err != nil {
		return nil, err
	}

	if err := s.transferRepo.CreateTransfer(ctx, transfer, senderRecord, recipientRecord); err != nil {
		return nil, err
	}
	dto := newPointTransferDTO(transfer)
	return &dto, nil
}
