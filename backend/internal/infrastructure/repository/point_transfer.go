package repository

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"jcourse/internal/domain/point"
)

func newPointTransferEntity(t *point.Transfer) PointTransferEntity {
	return PointTransferEntity{
		ID:              t.ID,
		SenderUserID:    t.SenderUserID,
		RecipientUserID: t.RecipientUserID,
		Amount:          t.Amount,
		Fee:             t.Fee,
		FeePayer:        string(t.FeePayer),
		SenderDelta:     t.SenderDelta,
		RecipientDelta:  t.RecipientDelta,
		CreatedAt:       t.CreatedAt,
	}
}

func newPointRecordEntity(r point.Record) UserPointRecordEntity {
	return UserPointRecordEntity{
		ID:          r.ID,
		UserID:      r.UserID,
		Reason:      string(r.Reason),
		Amount:      r.Amount,
		Description: r.Description,
		CreatedAt:   r.CreatedAt,
	}
}

func (r *PointRepository) CreateTransfer(ctx context.Context, t *point.Transfer, senderRecord point.Record, recipientRecord point.Record) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var sender UserEntity
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", t.SenderUserID).Take(&sender).Error; err != nil {
			return err
		}

		var balance int64
		if err := tx.Model(&UserPointRecordEntity{}).
			Where("user_id = ?", t.SenderUserID).
			Select("COALESCE(SUM(amount), 0)").
			Scan(&balance).Error; err != nil {
			return err
		}
		if balance+int64(t.SenderDelta) < 0 {
			return point.ErrInsufficientBalance
		}

		transferEntity := newPointTransferEntity(t)
		if err := tx.Create(&transferEntity).Error; err != nil {
			return err
		}
		t.ID = transferEntity.ID

		senderEntity := newPointRecordEntity(senderRecord)
		recipientEntity := newPointRecordEntity(recipientRecord)
		if err := tx.Create(&senderEntity).Error; err != nil {
			return err
		}
		return tx.Create(&recipientEntity).Error
	})
}

var _ point.TransferRepository = (*PointRepository)(nil)
