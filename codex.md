# Codex 项目交接

更新时间：2026-10-03，Asia/Shanghai。本文用于新任务窗口恢复上下文；进程、端口和工作区状态以新窗口实查为准。

## 用户目标与工作要求

- 项目原后端是 Python FastAPI，正在独立目录 `backend_go/` 中迁移为 Go，技术栈为 Gin、GORM、Redis。
- Go 接口继续兼容现有 Vue 前端、MySQL 表结构、用户密码哈希和登录令牌。
- 用户明确要求 Go 代码添加逐行中文注释，后续新增或修改 Go 代码也应保持这个要求。
- 用户明确指出过“model 全放在一个文件，而且没有分层”。这已完成整改：模型各自独立，实际请求经过 handler、service、repository。
- 使用中文沟通，直接完成明确需求并验证，避免只给计划或反复确认常规操作。
- 上一轮分层改造已完成，没有已知阻塞项；后续功能由用户的新需求决定，不自行扩展成另一轮重写。

## 工作区与环境

- 根目录：`D:\projects\news_py_go`；系统 Windows，终端 PowerShell。
- Git 分支核对时为 `main`，最近提交为 `a50b0fb add frontend and project documentation`。
- 当前大量改动未提交，`backend_go/` 整个目录仍未跟踪。新窗口必须先运行 `git status --short`，不要因未跟踪就重建或删除。
- 根目录 `.env.example` 在此前已有删除状态；除非用户要求，不自行恢复。
- 已有 `.venv/`、前端 `node_modules/` 和 Go 构建产物；先核对可用性，避免重复安装。
- 系统 Go 命令路径为 `C:\go1.25.0.windows-amd64\go\bin\go.exe`。系统版本为 1.25.0，模块声明 `go 1.26.0`，进入 `backend_go/` 后自动选择工具链；本次核对模块内为 Go 1.26.0。
- 本次核对 Node 为 24.19.0，npm 为 11.17.0。
- 实际配置在被 Git 忽略的根目录 `.env`；不要把数据库密码或 AI 密钥写入交接文档、代码、日志或提交。

## 已完成的工作

### Python 后端与前端修复

- 创建 Python 虚拟环境并安装依赖，清理重复的 python-dotenv 版本；uvloop 增加 Windows 平台条件。
- 修复新闻、收藏、历史响应字段，前端统一使用 `publishTime`；改善列表加载状态、分类请求竞争和历史删除行为。
- 历史删除按历史记录主键执行，并限定当前用户，不能把新闻 ID 当成历史 ID。
- 移除前端源码中的 AI 密钥，改成登录后访问后端 `/api/ai/chat`，后端代理 SSE。
- Python 新增 `routers/ai.py` 和 `tests/test_api.py`；前端新增 Vitest/jsdom 测试。
- 历史上前端曾包含 AI 密钥，README 已提示撤销旧密钥；是否已在服务商侧完成轮换未确认，不能声称已完成。

### Go 后端迁移

Go 模块：`github.com/prejudicing/news_py_go/backend_go`。

已实现新闻分类、列表、分页、详情、浏览量、相关新闻；注册、登录、认证、资料和密码修改；收藏全流程；浏览历史全流程；认证后的 AI 流式转发；Redis 缓存、降级、CORS、参数验证、统一响应和健康检查。

关键兼容行为：

- 普通 JSON 接口使用 `{code, message, data}`，首页例外；详情等响应字段兼容当前前端。
- 新闻查询参数使用 `categoryId`、`page`、`pageSize`；收藏和历史添加使用 JSON `newsId`。
- 时间字段使用 `publishTime`、`favoriteTime`、`viewTime`。
- 认证支持原有裸令牌和 `Bearer <token>`，沿用数据库 `user_token`。
- 登录事务锁定用户、删除旧令牌、签发新令牌，有效期 7 天；兼容 Python bcrypt 哈希及旧长密码截断行为。
- 新闻详情即使命中缓存，每次浏览仍递增数据库浏览量。
- 相关新闻按新闻所属分类查询并排除当前新闻，已有回归断言防止新闻 ID 与分类 ID 参数颠倒。
- 重复浏览同一新闻更新已有历史，保持历史记录 ID；删除始终限定当前用户。
- Go Redis 前缀默认 `go:news:`，避免与 Python 缓存格式冲突；Redis 不可用时回退到数据库查询。
- AI 请求传递取消上下文，禁止自动跟随重定向，统一隐藏上游错误细节；测试使用模拟上游，没有真实模型调用。
- 使用已有数据库，服务启动不执行 AutoMigrate，不修改业务表或导入样例数据。

### 已落地的分层结构

```text
backend_go/
  cmd/server/main.go           初始化 MySQL、Redis，启动和关闭服务
  internal/config/            环境配置、时区、MySQL DSN
  internal/model/             数据库实体，每个模型独立文件
    user.go
    user_token.go
    category.go
    news.go
    favorite.go
    history.go
  internal/dto/               请求参数、验证标签和响应结构
  internal/handler/           Gin 路由、参数绑定、认证中间件、响应和 SSE 转发
  internal/service/           登录、缓存、收藏、历史、AI 等业务规则
  internal/repository/        Store 接口、GORM 实现、SQL、行锁和错误转换
  internal/cache/             Redis JSON 缓存和降级
  internal/server/            依赖装配及 HTTP 集成测试
```

调用链：`handler → service → repository → MySQL`。Service 可使用 Redis 缓存封装。

- Handler 不直接访问 GORM、Redis 或执行密码哈希操作。
- Service 不引用 Gin、GORM 或 MySQL 驱动，通过 `repository.Store` 访问数据；业务层决定事务边界，仓储执行实际事务。
- Repository 负责持久化和数据库错误转换，向上暴露 `ErrNotFound`、`ErrDuplicate`、`ErrRelation`。
- DTO 与数据库 model 分开；SQL 联表查询结构留在 repository。
- `server.New(db, cache, cfg)` 保持装配入口，实例持有 `Service`，`Router()` 委托 handler 注册路由。
- 已移除旧 `internal/server/users.go`、`news.go`、`collections.go`、`ai.go`，不要重新放回单体控制器实现。
- 本次共有 37 个 Go 文件，全部非空代码行已有前置中文注释，包括测试；注释补全后用 Go 扫描器确认代码 token 未改变。

## 当前服务快照

以下是写本文时的核对结果，新窗口不能直接复用 PID：

| 服务 | 地址 | 核对结果 | 当时 PID |
| --- | --- | --- | --- |
| Go 后端 | http://127.0.0.1:8001 | `/health` 返回 200，MySQL `ok`、Redis `true` | 36668 |
| Vue 开发服务 | http://127.0.0.1:5173 | HTTP 200 | 15100 |
| Python 后端 | http://127.0.0.1:8000 | 端口在监听，此次未重测全部接口 | 38292 |

Go 运行文件：`backend_go/bin/news-server.exe`。根目录日志为 `backend_go.stdout.log`、`backend_go.stderr.log`；前端日志为 `frontend.stdout.log`、`frontend.stderr.log`。二进制和日志已忽略。

前端路径必须保留原拼写 `fronted/xwzx-news/`。该目录 `.env.local` 当前配置 `VITE_API_BASE_URL='http://127.0.0.1:8001'`；源码默认回退仍是 8000。

本机 5173 还存在其他地址上的监听进程，管理服务时以 `127.0.0.1` 对应的项目进程为准，不要仅凭端口批量终止进程。

## 启动与验证命令

以下命令按所在目录分别执行。服务已运行时先检查，不要直接启动第二个实例。

Go 启动：

```powershell
Set-Location D:\projects\news_py_go\backend_go
go run ./cmd/server
```

默认监听 8001，配置优先级为已有环境变量、目录内 `.env`、根目录 `.env`；设置 `ENV_FILE` 时加载指定文件。可用 `GO_HTTP_ADDR` 修改监听地址。

Go 回归验证：

```powershell
Set-Location D:\projects\news_py_go\backend_go
go test ./...
go vet ./...
$env:ENV_FILE = (Resolve-Path ../.env).Path
$env:GO_INTEGRATION_TEST = '1'
go test ./... -count=1
```

集成测试仅在 `GO_INTEGRATION_TEST=1` 时运行，否则会跳过。它按原 SQL 建立 `news_go_test_<数字>` 临时数据库，并在结束时删除该测试库；需要当前账号具备建库和删库权限，不应改写测试去操作业务库。Redis 使用 miniredis，AI 使用本地模拟 HTTP 服务。

Windows 下运行中的 exe 被占用。需要更新时先构建到另一文件，例如：

```powershell
go build -o bin/news-server-next.exe ./cmd/server
```

构建成功后核对监听 PID 和可执行文件绝对路径，只停止本项目 Go 进程，再替换原 exe 并启动。用 `Start-Process` 启动后台服务时带 `-WindowStyle Hidden`，指定工作目录和日志路径；最后复查 `/health`。

前端启动：

```powershell
Set-Location D:\projects\news_py_go\fronted\xwzx-news
npm run dev -- --host 127.0.0.1 --port 5173 --strictPort
```

前端验证（Go 服务需要先正常运行）：

```powershell
Set-Location D:\projects\news_py_go\fronted\xwzx-news
$env:FRONTEND_LIVE_TEST = '1'
npm test
npm run build
```

Python 对照验证（仅在需要修改或比较 Python 时执行）：

```powershell
Set-Location D:\projects\news_py_go
.\.venv\Scripts\python.exe -m unittest discover -s tests -v
.\.venv\Scripts\python.exe -m uvicorn main:app --host 127.0.0.1 --port 8000
```

## 已有验证证据

上一轮完成代码后执行并通过：

- Go 全包测试、真实 MySQL 临时库集成测试、`go vet ./...`、新二进制构建。
- Service 独立测试覆盖详情缓存命中仍计数、正确的相关新闻参数、历史事务内新增/更新及写入错误传递。
- HTTP 集成测试覆盖注册登录与令牌轮换、Python 哈希/令牌兼容、资料密码、新闻分页/缓存/浏览量、收藏、历史主键/用户隔离、AI SSE 与错误脱敏。
- 前端 4 个测试文件共 10 项通过，包含连接实际运行 Go 服务的首页渲染测试；`npm run build` 通过。
- 新 Go 二进制已重启，健康检查为 MySQL 正常、Redis 正常；前端 HTTP 200。
- Python 7 项测试在此前修复阶段通过；最新分层改造阶段没有重新跑 Python 测试。

前端验证为 Vitest/jsdom 和 HTTP 联通，尚未做人工浏览器视觉验收，不能把它描述为所有页面均已视觉验证。

## 新窗口建议读取顺序

1. 本文、根目录 `README.md`、`backend_go/README.md`，然后查看 `git status --short`。
2. `backend_go/cmd/server/main.go`、`internal/server/server.go`、`internal/handler/handler.go`，了解实际依赖和路由。
3. 根据新需求阅读对应的 dto、model、handler、service、repository 及测试，不必重新逐一实现所有模块。
4. 接口与表结构参考 `docs/01-接口规范文档/API接口规范文档.md`、`docs/02-数据库sql文件/database.sql`、`docs/项目后端设计说明文档.md`；同时核对前端实际调用，文档与实现冲突时不要凭猜测改接口。
5. 核对端口、进程、健康检查和前端实际 API 地址，再依据用户的新需求继续。

保留当前未提交改动；不要执行 `git reset --hard`、`git clean` 或覆盖用户文件来恢复所谓干净状态。未经用户要求，不提交、推送或操作远程仓库。

## 给新窗口的提示词

```text
请接手 D:\projects\news_py_go 项目。先阅读根目录 codex.md、README.md 和 backend_go/README.md，再检查 Git 工作区及实际运行状态。

我们已把 FastAPI 后端迁移到 backend_go，使用 Gin、GORM、Redis，完成 handler → service → repository 分层、独立 model 文件和 dto，Go 代码有逐行中文注释。继续开发时保持这些要求，并兼容现有 Vue 前端、数据库表、密码哈希和令牌。

保留所有未提交和未跟踪的改动，不要重复搭建已完成模块或恢复已删除的旧单体实现。配置从现有 .env 读取，不输出密钥。管理服务前核对 PID 和可执行路径。

先简短说明当前状态，再按我接下来给出的具体需求直接完成修改、相应测试和必要的服务验证。验证结果区分实际接口联通、自动化渲染测试与浏览器视觉检查，不把未执行的检查写成通过。

本次具体需求：[在这里填写新任务，也可以在下一条消息发送。]
```
