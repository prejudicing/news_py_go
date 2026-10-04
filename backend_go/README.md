# Go 后端

现有 FastAPI 后端的 Go 实现，使用 Gin、GORM 和 go-redis。

已实现：

- 新闻分类、列表、分页、详情、浏览量与相关新闻
- 注册、登录、令牌认证、用户资料和密码修改
- 收藏状态、添加、删除、列表与清空
- 浏览历史添加、更新、列表、按历史记录 ID 删除与清空
- 登录后的 AI 流式转发，密钥只在后端读取
- Redis 缓存、CORS、参数验证、统一错误响应与健康检查

## 启动

需要 MySQL、Redis 和 Go 1.26+。模块声明 Go 1.26，支持自动工具链下载的 Go 安装会自动下载所需版本。

在项目根目录执行：

```powershell
cd backend_go
go run ./cmd/server
```

默认监听 http://127.0.0.1:8001，读取 `backend_go/.env` 与项目根目录 `.env`，已有环境变量优先，目录内 `.env` 优先于根目录。可用 `ENV_FILE` 指定配置文件，或用 `GO_HTTP_ADDR` 修改地址。

```powershell
$env:GO_HTTP_ADDR = '127.0.0.1:8001'
go build -o bin/news-server.exe ./cmd/server
.\bin\news-server.exe
```

服务沿用现有 `news_app` 数据库，不会自动建表、修改表或导入样例数据。首次部署需按照项目根目录 `docs/02-数据库sql文件/database.sql` 初始化数据库。

健康检查：http://127.0.0.1:8001/health。返回 MySQL 与 Redis 连接状态。Redis 不可用时新闻查询回退到 MySQL。

## 接口

接口路径和主要响应字段与当前前端兼容。除首页外，JSON 响应采用 `{code, message, data}`。认证支持原有裸令牌和 `Bearer <token>`。

| 方法 | 路径 | 认证 |
| --- | --- | --- |
| GET | `/api/news/categories`、`/api/news/list`、`/api/news/detail` | 否 |
| POST | `/api/user/register`、`/api/user/login` | 否 |
| GET | `/api/user/info` | 是 |
| PUT | `/api/user/update`、`/api/user/password` | 是 |
| GET | `/api/favorite/check`、`/api/favorite/list` | 是 |
| POST | `/api/favorite/add` | 是 |
| DELETE | `/api/favorite/remove`、`/api/favorite/clear` | 是 |
| POST | `/api/history/add` | 是 |
| GET | `/api/history/list` | 是 |
| DELETE | `/api/history/delete/:history_id`、`/api/history/clear` | 是 |
| POST | `/api/ai/chat` | 是 |

新闻列表使用 `categoryId`、`page`、`pageSize`，详情使用 `id`；收藏和历史添加使用 JSON `newsId`。时间字段统一为 `publishTime`、`favoriteTime` 和 `viewTime`。

原有 bcrypt 密码哈希和数据库令牌可继续使用。新登录会更新该用户令牌，有效期为 7 天。Go 使用独立的 `go:news:` Redis 前缀，避免与 Python 的缓存格式冲突。历史删除按历史记录 ID 执行，并限定当前用户。

AI 配置继续使用根目录 `.env` 中的 `AI_API_KEY`、`AI_MODEL` 和 `AI_API_ENDPOINT`。AI 测试使用本地模拟服务，不产生真实模型调用。

## 前端连接

在 `fronted/xwzx-news/.env.local` 中设置：

```dotenv
VITE_API_BASE_URL=http://127.0.0.1:8001
```

然后运行 `npm run dev -- --host 127.0.0.1 --port 5173`。前端地址为 http://127.0.0.1:5173。若需切回 Python，把后端地址改回 `http://127.0.0.1:8000`。

## 目录

```text
cmd/server/         服务入口、连接初始化和退出处理
internal/config/    环境配置和 MySQL DSN
internal/model/     每张表独立一个模型文件
internal/dto/       HTTP 请求参数和响应结构
internal/handler/   按路由、中间件、参数和响应拆分的 HTTP 适配层
internal/service/   用户、新闻、收藏、历史、AI 和健康检查领域服务
internal/repository/ 按领域拆分的仓储接口、GORM 实现和错误转换
internal/cache/     Redis JSON 缓存及降级
internal/server/    组装依赖和接口集成测试
```

请求调用顺序为 `handler → service → repository → MySQL`。业务服务拆分为 `UserService`、`NewsService`、`FavoriteService`、`HistoryService`、`AIService` 和 `HealthService`，每个服务只依赖对应的领域仓储接口。登录与历史更新分别使用各自的事务入口；仓储实现负责事务执行、行锁和 SQL。Handler 不直接操作 GORM 或 Redis，Service 不引用 Gin 或 GORM。

模型分为 `user.go`、`user_token.go`、`category.go`、`news.go`、`favorite.go` 和 `history.go`。DTO 单独定义前端字段，数据库模型不再同时承担接口响应职责。注释用于说明职责、约束和非直观设计原因，不再逐行复述代码。

## 测试

在 `backend_go` 目录执行：

```powershell
go test ./...
go vet ./...
```

完整 MySQL 集成测试会创建一个 `news_go_test_<数字>` 临时数据库，使用原 SQL 文件的建表语句验证兼容性，并在结束时删除该测试库；需要数据库账号具备建库和删库权限。不会修改配置中的业务数据库。Redis 和 AI 使用本地模拟服务。

```powershell
$env:ENV_FILE = (Resolve-Path ../.env).Path
$env:GO_INTEGRATION_TEST = '1'
go test ./... -count=1
```

Go 服务运行时，在前端目录执行联通测试：

```powershell
$env:FRONTEND_LIVE_TEST = '1'
npm test
npm run build
```

框架资料：[Gin](https://gin-gonic.com/en/docs/)、[GORM MySQL](https://gorm.io/docs/connecting_to_the_database.html)、[go-redis](https://redis.io/docs/latest/develop/clients/go/)。
