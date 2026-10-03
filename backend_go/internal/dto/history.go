// 声明接口请求和响应结构包。
package dto

// 定义历史列表响应。
type HistoryItem struct {
	// 嵌入前端新闻列表字段，复用新闻响应格式。
	NewsItem
	// 保存浏览时间；JSON 标签定义接口字段名。
	ViewTime string `json:"viewTime"`
	// 保存历史记录主键；JSON 标签定义接口字段名。
	HistoryID uint64 `json:"historyId"`
	// 结束当前声明或代码作用域。
}
