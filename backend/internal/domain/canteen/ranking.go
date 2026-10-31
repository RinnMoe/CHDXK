package canteen

const TaskTypeRefreshCanteenRankings = "canteen:refresh_rankings"

// RefreshRankingsTask recomputes persisted store stats and clears canteen
// caches. It mirrors the course rating refresh task.
type RefreshRankingsTask struct{}

func NewRefreshRankingsTask() RefreshRankingsTask {
	return RefreshRankingsTask{}
}

func (t RefreshRankingsTask) Type() string {
	return TaskTypeRefreshCanteenRankings
}

func (t RefreshRankingsTask) Payload() []byte {
	return nil
}
