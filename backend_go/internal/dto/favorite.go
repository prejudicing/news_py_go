// 声明接口请求和响应结构包。
package dto

// 定义包含新闻主键的请求。
type NewsRequest struct {
	// 保存新闻主键；JSON 标签定义接口字段名，binding 标签约束输入。
	NewsID uint64 `json:"newsId" binding:"required,gt=0"`
	// 结束当前声明或代码作用域。
}

// 定义收藏列表响应。
type FavoriteItem struct {
	// 嵌入前端新闻列表字段，复用新闻响应格式。
	NewsItem
	// 保存收藏时间；JSON 标签定义接口字段名。
	FavoriteTime string `json:"favoriteTime"`
	// 保存收藏记录主键；JSON 标签定义接口字段名。
	FavoriteID uint64 `json:"favoriteId"`
	// 结束当前声明或代码作用域。
}
