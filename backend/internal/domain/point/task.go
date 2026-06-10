package point

import "encoding/json"

const (
	TaskTypeGrantReward             = "point:grant_reward"
	TaskTypeRevokeReviewRewardsByID = "point:revoke_review_rewards_by_id"
)

type GrantRewardPayload struct {
	RewardID int `json:"reward_id"`
}

type RevokeReviewRewardsByIDPayload struct {
	ReviewID     int `json:"review_id"`
	CourseID     int `json:"course_id"`
	AuthorUserID int `json:"author_user_id"`
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

type RevokeReviewRewardsByIDTask struct {
	payload RevokeReviewRewardsByIDPayload
}

func NewRevokeReviewRewardsByIDTask(reviewID, courseID, authorUserID int) RevokeReviewRewardsByIDTask {
	return RevokeReviewRewardsByIDTask{payload: RevokeReviewRewardsByIDPayload{ReviewID: reviewID, CourseID: courseID, AuthorUserID: authorUserID}}
}

func (t RevokeReviewRewardsByIDTask) Type() string {
	return TaskTypeRevokeReviewRewardsByID
}

func (t RevokeReviewRewardsByIDTask) Payload() []byte {
	b, _ := json.Marshal(t.payload)
	return b
}
