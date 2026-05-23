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
	RecipientUsername string         `json:"recipient_username"`
	Amount            int            `json:"amount"`
	FeePayer          point.FeePayer `json:"fee_payer"`
}

type PointCommandService struct {
	accountRepo     account.AccountRepository
	transferRepo    point.TransferRepository
	transferService *point.TransferService
}

func NewPointCommandService(accountRepo account.AccountRepository, transferRepo point.TransferRepository, transferService *point.TransferService) *PointCommandService {
	return &PointCommandService{
		accountRepo:     accountRepo,
		transferRepo:    transferRepo,
		transferService: transferService,
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

	recipient, err := s.accountRepo.FindByUsername(ctx, strings.TrimSpace(cmd.RecipientUsername))
	if err != nil {
		return nil, err
	}
	if recipient == nil {
		return nil, ErrPointTransferRecipientNotFound
	}

	now := time.Now()
	transfer, senderRecord, recipientRecord, err := s.transferService.NewTransfer(
		point.UserRef{ID: sender.ID, Username: senderAccount.Username},
		point.UserRef{ID: recipient.ID, Username: recipient.Username},
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
