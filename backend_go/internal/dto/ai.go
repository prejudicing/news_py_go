// 声明接口请求和响应结构包。
package dto

// 定义单条对话消息。
type ChatMessage struct {
	// 保存对话角色；JSON 标签定义接口字段名，binding 标签约束输入。
	Role string `json:"role" binding:"required,oneof=user assistant"`
	// 保存正文或对话内容；JSON 标签定义接口字段名，binding 标签约束输入。
	Content string `json:"content" binding:"required,max=20000"`
	// 结束当前声明或代码作用域。
}

// 定义AI 对话请求。
type ChatRequest struct {
	// 保存对话消息列表；JSON 标签定义接口字段名，binding 标签约束输入。
	Messages []ChatMessage `json:"messages" binding:"required,min=1,max=50,dive"`
	// 结束当前声明或代码作用域。
}
