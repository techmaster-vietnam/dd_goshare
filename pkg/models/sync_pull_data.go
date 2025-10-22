package models

type SyncPullDataItem struct {
	Streak                *int                    `json:"streak"`
	Score                 *int                    `json:"score"`
	TotalDialogsCompleted *int                    `json:"total_dialogs_completed"`
	CustomerTopicProgress []CustomerTopicProgress `json:"customerTopicProgressTable"`
	DialogCompletion      []DialogCompletion      `json:"dialogCompletionTable"`
}
type SyncPullDataRequest struct {
	Data SyncPullDataItem `json:"data"`
}

type SyncPullDataResponse struct {
	StatusCode int              `json:"status_code"`
	Message    string           `json:"message"`
	Data       SyncPullDataItem `json:"data"`
}
