package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/point"
)

var (
	ErrPointTransferInvalidAmount        = point.ErrTransferInvalidAmount
	ErrPointTransferInvalidFeePayer      = point.ErrTransferInvalidFeePayer
	ErrPointTransferSelf                 = errors.New("cannot transfer points to self")
	ErrPointTransferRecipientNotFound    = errors.New("point transfer recipient not found")
	ErrPointTransferRecipientAmountSmall = point.ErrTransferRecipientAmountSmall
)

type PointTransferFeeConfig = point.TransferFeeConfig

type CreatePointTransferCommand struct {
	RecipientUsername string         `json:"recipient_username"`
	Amount            int            `json:"amount"`
	FeePayer          point.FeePayer `json:"fee_payer"`
}

type PointCommandService struct {
	userRepo        auth.UserRepository
	transferRepo    point.TransferRepository
	transferService *point.TransferService
}

func NewPointCommandService(userRepo auth.UserRepository, transferRepo point.TransferRepository, feeConfig PointTransferFeeConfig) *PointCommandService {
	return &PointCommandService{
		userRepo:        userRepo,
		transferRepo:    transferRepo,
		transferService: point.NewTransferService(feeConfig),
	}
}

func (s *PointCommandService) CreateTransfer(ctx context.Context, sender *auth.User, cmd CreatePointTransferCommand) (*PointTransferDTO, error) {
	recipient, err := s.userRepo.FindByUsername(ctx, strings.TrimSpace(cmd.RecipientUsername))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPointTransferRecipientNotFound
		}
		return nil, err
	}
	if recipient.ID == sender.ID {
		return nil, ErrPointTransferSelf
	}

	now := time.Now()
	transfer, senderRecord, recipientRecord, err := s.transferService.NewTransfer(
		point.UserRef{ID: sender.ID, Username: sender.Username},
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

func (s *PointCommandService) PreviewTransfer(cmd CreatePointTransferCommand) (*PointTransferPreviewDTO, error) {
	preview, err := s.transferService.Preview(cmd.Amount, cmd.FeePayer)
	if err != nil {
		return nil, err
	}
	dto := newPointTransferPreviewDTO(preview)
	return &dto, nil
}
