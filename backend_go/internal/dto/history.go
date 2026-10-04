package dto

// 定义历史列表响应。
type HistoryItem struct {
	NewsItem
	ViewTime  string `json:"viewTime"`
	HistoryID uint64 `json:"historyId"`
}
