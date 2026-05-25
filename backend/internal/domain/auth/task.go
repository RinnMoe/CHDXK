package auth

import (
	"encoding/json"
	"time"
)

const (
	TaskTypeClearExpiredSuspension = "auth:clear_expired_suspension"
	TaskTypeFlushAccess            = "auth:flush_access"
	TaskTypeSuspendUser            = "auth:suspend_user"
)

type ClearExpiredSuspensionPayload struct {
	UserID int `json:"user_id"`
}

type ClearExpiredSuspensionTask struct {
	payload ClearExpiredSuspensionPayload
}

func NewClearExpiredSuspensionTask(userID int) ClearExpiredSuspensionTask {
	return ClearExpiredSuspensionTask{payload: ClearExpiredSuspensionPayload{UserID: userID}}
}

func (t ClearExpiredSuspensionTask) Type() string {
	return TaskTypeClearExpiredSuspension
}

func (t ClearExpiredSuspensionTask) Payload() []byte {
	b, _ := json.Marshal(t.payload)
	return b
}

type FlushAccessPayload struct{}

type FlushAccessTask struct {
	payload FlushAccessPayload
}

func NewFlushAccessTask() FlushAccessTask {
	return FlushAccessTask{payload: FlushAccessPayload{}}
}

func (t FlushAccessTask) Type() string {
	return TaskTypeFlushAccess
}

func (t FlushAccessTask) Payload() []byte {
	b, _ := json.Marshal(t.payload)
	return b
}

type SuspendUserPayload struct {
	UserID   int           `json:"user_id"`
	Duration time.Duration `json:"duration"`
}

type SuspendUserTask struct {
	payload SuspendUserPayload
}

func NewSuspendUserTask(userID int, duration time.Duration) SuspendUserTask {
	return SuspendUserTask{payload: SuspendUserPayload{UserID: userID, Duration: duration}}
}

func (t SuspendUserTask) Type() string {
	return TaskTypeSuspendUser
}

func (t SuspendUserTask) Payload() []byte {
	b, _ := json.Marshal(t.payload)
	return b
}
