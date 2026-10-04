package dto

// 定义单条对话消息。
type ChatMessage struct {
	Role    string `json:"role" binding:"required,oneof=user assistant"`
	Content string `json:"content" binding:"required,max=20000"`
}

// 定义AI 对话请求。
type ChatRequest struct {
	Messages []ChatMessage `json:"messages" binding:"required,min=1,max=50,dive"`
}
