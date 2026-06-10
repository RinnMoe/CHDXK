//go:build test

package point

import (
	"context"
	"time"
)

type MockTransferRepository struct {
	Transfer        *Transfer
	SenderRecord    Record
	RecipientRecord Record

	OnCreateTransfer func(context.Context, *Transfer, Record, Record) error
}

func (r *MockTransferRepository) CreateTransfer(ctx context.Context, t *Transfer, senderRecord Record, recipientRecord Record) error {
	if r.OnCreateTransfer != nil {
		return r.OnCreateTransfer(ctx, t, senderRecord, recipientRecord)
	}
	copy := *t
	if copy.ID == 0 {
		copy.ID = 99
		t.ID = copy.ID
	}
	r.Transfer = &copy
	r.SenderRecord = senderRecord
	r.RecipientRecord = recipientRecord
	return nil
}

type MockRewardRepository struct {
	GrantedRewardID int
	GrantedAt       time.Time
	RevokedSources  []RewardSource
	RevokedAt       time.Time

	OnGrantReward            func(context.Context, int, time.Time) error
	OnRevokeRewardsBySources func(context.Context, []RewardSource, time.Time) error
}

func (r *MockRewardRepository) GrantReward(ctx context.Context, rewardID int, now time.Time) error {
	if r.OnGrantReward != nil {
		return r.OnGrantReward(ctx, rewardID, now)
	}
	r.GrantedRewardID = rewardID
	r.GrantedAt = now
	return nil
}

func (r *MockRewardRepository) RevokeRewardsBySources(ctx context.Context, sources []RewardSource, now time.Time) error {
	if r.OnRevokeRewardsBySources != nil {
		return r.OnRevokeRewardsBySources(ctx, sources, now)
	}
	r.RevokedSources = append([]RewardSource(nil), sources...)
	r.RevokedAt = now
	return nil
}
