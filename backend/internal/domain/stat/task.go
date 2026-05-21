package stat

import "encoding/json"

const TaskTypeCollectDailySiteStats = "stat:collect_daily_site_stats"

type CollectDailySiteStatsPayload struct {
	StatDate string `json:"stat_date,omitempty"`
}

type CollectDailySiteStatsTask struct {
	payload CollectDailySiteStatsPayload
}

func NewCollectDailySiteStatsTask(statDate string) CollectDailySiteStatsTask {
	return CollectDailySiteStatsTask{payload: CollectDailySiteStatsPayload{StatDate: statDate}}
}

func (t CollectDailySiteStatsTask) Type() string {
	return TaskTypeCollectDailySiteStats
}

func (t CollectDailySiteStatsTask) Payload() []byte {
	b, _ := json.Marshal(t.payload)
	return b
}
