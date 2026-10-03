// 声明接口请求和响应结构包。
package dto

// NewsItem 定义前端使用的新闻字段，与数据库模型分开。
type NewsItem struct {
	// 保存记录主键；JSON 标签定义接口字段名。
	ID uint64 `json:"id"`
	// 保存新闻标题；JSON 标签定义接口字段名。
	Title string `json:"title"`
	// 保存新闻摘要；JSON 标签定义接口字段名。
	Description *string `json:"description"`
	// 保存新闻图片地址；JSON 标签定义接口字段名。
	Image *string `json:"image"`
	// 保存新闻作者；JSON 标签定义接口字段名。
	Author *string `json:"author"`
	// 保存所属分类主键；JSON 标签定义接口字段名。
	CategoryID uint64 `json:"categoryId"`
	// 保存浏览次数；JSON 标签定义接口字段名。
	Views int64 `json:"views"`
	// 保存发布时间；JSON 标签定义接口字段名。
	PublishTime string `json:"publishTime"`
	// 结束当前声明或代码作用域。
}

// 定义相关新闻摘要响应。
type RelatedItem struct {
	// 保存记录主键；JSON 标签定义接口字段名。
	ID uint64 `json:"id"`
	// 保存新闻标题；JSON 标签定义接口字段名。
	Title string `json:"title"`
	// 保存新闻图片地址；JSON 标签定义接口字段名。
	Image *string `json:"image"`
	// 保存浏览次数；JSON 标签定义接口字段名。
	Views int64 `json:"views"`
	// 结束当前声明或代码作用域。
}

// 定义新闻详情响应。
type NewsDetail struct {
	// 保存记录主键；JSON 标签定义接口字段名。
	ID uint64 `json:"id"`
	// 保存新闻标题；JSON 标签定义接口字段名。
	Title string `json:"title"`
	// 保存正文或对话内容；JSON 标签定义接口字段名。
	Content string `json:"content"`
	// 保存新闻图片地址；JSON 标签定义接口字段名。
	Image *string `json:"image"`
	// 保存新闻作者；JSON 标签定义接口字段名。
	Author *string `json:"author"`
	// 保存发布时间；JSON 标签定义接口字段名。
	PublishTime string `json:"publishTime"`
	// 保存所属分类主键；JSON 标签定义接口字段名。
	CategoryID uint64 `json:"categoryId"`
	// 保存浏览次数；JSON 标签定义接口字段名。
	Views int64 `json:"views"`
	// 保存相关新闻列表；JSON 标签定义接口字段名。
	RelatedNews []RelatedItem `json:"relatedNews"`
	// 结束当前声明或代码作用域。
}

// 定义通用分页响应。
type Page[T any] struct {
	// 保存当前页记录列表；JSON 标签定义接口字段名。
	List []T `json:"list"`
	// 保存符合条件的总记录数；JSON 标签定义接口字段名。
	Total int64 `json:"total"`
	// 保存是否还有下一页；JSON 标签定义接口字段名。
	HasMore bool `json:"hasMore"`
	// 结束当前声明或代码作用域。
}
