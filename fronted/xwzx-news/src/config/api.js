// 可使用 VITE_API_BASE_URL 指定其他后端地址。
export const apiConfig = {
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://127.0.0.1:8000',
}

// AI 凭据由后端管理，前端只访问本项目的接口。
export const aiChatConfig = {
  apiEndpoint: `${apiConfig.baseURL}/api/ai/chat`,
}
