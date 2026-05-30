package point

import "encoding/json"

const TaskTypeGrantReward = "point:grant_reward"

type GrantRewardPayload struct {
	RewardID int `json:"reward_id"`
}

type GrantRewardTask struct {
	payload GrantRewardPayload
}

func NewGrantRewardTask(rewardID int) GrantRewardTask {
	return GrantRewardTask{payload: GrantRewardPayload{RewardID: rewardID}}
}

func (t GrantRewardTask) Type() string {
	return TaskTypeGrantReward
}

func (t GrantRewardTask) Payload() []byte {
	b, _ := json.Marshal(t.payload)
	return b
}
