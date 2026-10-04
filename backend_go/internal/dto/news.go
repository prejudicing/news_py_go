package dto

// NewsItem 定义前端使用的新闻字段，与数据库模型分开。
type NewsItem struct {
	ID          uint64  `json:"id"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Image       *string `json:"image"`
	Author      *string `json:"author"`
	CategoryID  uint64  `json:"categoryId"`
	Views       int64   `json:"views"`
	PublishTime string  `json:"publishTime"`
}

// 定义相关新闻摘要响应。
type RelatedItem struct {
	ID    uint64  `json:"id"`
	Title string  `json:"title"`
	Image *string `json:"image"`
	Views int64   `json:"views"`
}

// 定义新闻详情响应。
type NewsDetail struct {
	ID          uint64        `json:"id"`
	Title       string        `json:"title"`
	Content     string        `json:"content"`
	Image       *string       `json:"image"`
	Author      *string       `json:"author"`
	PublishTime string        `json:"publishTime"`
	CategoryID  uint64        `json:"categoryId"`
	Views       int64         `json:"views"`
	RelatedNews []RelatedItem `json:"relatedNews"`
}

// 定义通用分页响应。
type Page[T any] struct {
	List    []T   `json:"list"`
	Total   int64 `json:"total"`
	HasMore bool  `json:"hasMore"`
}
