package dto

// 定义包含新闻主键的请求。
type NewsRequest struct {
	NewsID uint64 `json:"newsId" binding:"required,gt=0"`
}

// 定义收藏列表响应。
type FavoriteItem struct {
	NewsItem
	FavoriteTime string `json:"favoriteTime"`
	FavoriteID   uint64 `json:"favoriteId"`
}
