package dto

// 定义历史列表响应。
type HistoryItem struct {
	NewsItem
	// ViewTime 是用户最近一次浏览该新闻的时间。
	ViewTime string `json:"viewTime"`
	// HistoryID 是历史记录自身的主键，不是新闻 ID。
	HistoryID uint64 `json:"historyId"`
}
