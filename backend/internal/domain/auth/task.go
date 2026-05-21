package auth

import (
	"encoding/json"
)

const TaskTypeClearExpiredSuspension = "auth:clear_expired_suspension"

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
