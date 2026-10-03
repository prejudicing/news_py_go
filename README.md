# news_py_go

新闻资讯项目，后端使用 FastAPI，前端使用 Vue 3 和 Vite。

Go 后端已实现于 [`backend_go`](backend_go/README.md)，使用 Gin、GORM 和 Redis，默认地址 http://127.0.0.1:8001，兼容现有前端与数据库。

## 后端启动（Windows）

```powershell
python -m venv .venv
uv pip install --python .venv/Scripts/python.exe -r requirements.txt
.\.venv\Scripts\python.exe -m uvicorn main:app --host 127.0.0.1 --port 8000
```

MySQL 和 Redis 连接配置放在项目根目录的 `.env`，接口文档地址为 http://127.0.0.1:8000/docs。

## 前端启动

```powershell
cd fronted/xwzx-news
npm ci
npm run dev -- --host 127.0.0.1 --port 5173
```

访问 http://127.0.0.1:5173。前端默认连接 http://127.0.0.1:8000，可通过前端目录的 `.env.local` 设置 `VITE_API_BASE_URL`。

## AI 问答

登录后使用 AI 问答。请求通过 `/api/ai/chat` 转发，密钥仅在后端读取。项目根目录 `.env` 支持以下配置：

```dotenv
AI_API_KEY=替换为新的密钥
AI_MODEL=qwen3-max-preview
AI_API_ENDPOINT=https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions
```

请在服务商控制台撤销曾提交到前端源码的密钥，再配置新密钥并重启后端。不要将密钥写入任何 `VITE_` 环境变量，这些变量会进入浏览器产物。

## 验证

```powershell
.\.venv\Scripts\python.exe -m unittest discover -s tests -v
cd fronted/xwzx-news
npm run build
npm test
```

AI 测试使用模拟上游，不产生真实模型调用。

后端运行时，可执行前端与后端联通的首页渲染测试：

```powershell
cd fronted/xwzx-news
$env:FRONTEND_LIVE_TEST = '1'
npm test
```
