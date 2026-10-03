// 声明业务接口与路由包。
package server

// 开始声明本文件使用的标准库、第三方库和项目内部包。
import (
	// 引入 bytes，用于将字节数据包装为可读取的请求体。
	"bytes"
	// 引入 database/sql，用于管理集成测试的底层数据库连接。
	"database/sql"
	// 引入 encoding/json，用于编码和解码 JSON 数据。
	"encoding/json"
	// 引入 fmt，用于格式化文本和构造错误。
	"fmt"
	// 引入 net/http，用于构造 HTTP 请求、客户端和服务。
	"net/http"
	// 引入 net/http/httptest，用于在测试中模拟 HTTP 请求、响应和上游服务。
	"net/http/httptest"
	// 引入 os，用于读取环境变量、配置文件或进程信号。
	"os"
	// 引入 regexp，用于校验测试库名称并提取建表语句。
	"regexp"
	// 引入 strings，用于处理令牌和响应头中的字符串。
	"strings"
	// 引入 testing，用于编写和运行 Go 测试。
	"testing"
	// 引入 time，用于处理时间、有效期和超时。
	"time"

	// 引入 github.com/alicebob/miniredis/v2，用于启动内存 Redis 测试服务。
	"github.com/alicebob/miniredis/v2"
	// 引入 github.com/gin-gonic/gin，用于使用 Gin 路由、请求上下文和 JSON 响应。
	"github.com/gin-gonic/gin"
	// 引入项目内部的 cache 包，复用项目 Redis 缓存封装。
	"github.com/prejudicing/news_py_go/backend_go/internal/cache"
	// 引入项目内部的 config 包，复用项目环境配置。
	"github.com/prejudicing/news_py_go/backend_go/internal/config"
	// 引入github.com/prejudicing/news_py_go/backend_go/internal/dto，用于接口参数和响应结构。
	"github.com/prejudicing/news_py_go/backend_go/internal/dto"
	// 引入项目内部的 model 包，复用数据库模型。
	"github.com/prejudicing/news_py_go/backend_go/internal/model"
	// 引入 github.com/redis/go-redis/v9，用于使用 Redis 客户端读取和写入缓存。
	"github.com/redis/go-redis/v9"
	// 引入 gorm.io/driver/mysql，用于使用 GORM 的 MySQL 适配器。
	"gorm.io/driver/mysql"
	// 引入 gorm.io/gorm，用于使用 GORM 执行数据库查询和事务。
	"gorm.io/gorm"
	// 引入 gorm.io/gorm/logger，用于配置 GORM 日志级别。
	"gorm.io/gorm/logger"
	// 结束依赖导入列表。
)

// 封装集成测试的请求发送与响应断言。
type testAPI struct {
	// 保存 当前测试对象。
	t *testing.T
	// 保存 被测试的 HTTP 路由。
	router http.Handler
	// 保存 请求使用的登录令牌。
	token string
	// 结束当前作用域的代码块或结构体定义。
}

// 发送测试请求并检查 HTTP 状态与响应格式。
func (a testAPI) request(method, path string, data any, status int) map[string]any {
	// 标记为测试辅助函数，让失败位置指向调用者。
	a.t.Helper()
	// 把数据编码成 JSON，供请求或缓存使用。
	body, _ := json.Marshal(data)
	// 构造发往被测接口的 HTTP 请求。
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	// 设置测试请求的认证头或 JSON 内容类型。
	req.Header.Set("Content-Type", "application/json")
	// 检查测试结果是否满足当前用例的预期条件。
	if a.token != "" {
		// 设置测试请求的认证头或 JSON 内容类型。
		req.Header.Set("Authorization", a.token)
		// 结束当前作用域的代码块或结构体定义。
	}
	// 创建测试响应记录器，接收状态码、响应头和响应体。
	w := httptest.NewRecorder()
	// 通过真实 Gin 路由执行测试请求。
	a.router.ServeHTTP(w, req)
	// 比较接口返回状态码与调用辅助器指定的期望值。
	if w.Code != status {
		// 记录失败原因并立即结束当前测试。
		a.t.Fatalf("%s %s: got %d want %d: %s", method, path, w.Code, status, w.Body.String())
		// 结束当前作用域的代码块或结构体定义。
	}
	// 声明变量，用于接收测试响应的完整 JSON 数据。
	var result map[string]any
	// 把 JSON 数据解码到目标对象，并检查解码错误。
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		// 记录失败原因并立即结束当前测试。
		a.t.Fatal(err)
		// 结束当前作用域的代码块或结构体定义。
	}
	// 确认 JSON 业务状态码与 HTTP 状态码一致。
	if result["code"] != float64(status) {
		// 记录失败原因并立即结束当前测试。
		a.t.Fatal("response envelope mismatch")
		// 结束当前作用域的代码块或结构体定义。
	}
	// 响应 data 为对象时提取该对象，便于后续字段断言。
	if value, valid := result["data"].(map[string]any); valid {
		// 返回已经取得的配置值或测试数据对象。
		return value
		// 结束当前作用域的代码块或结构体定义。
	}
	// 未提取到对象形式的 data 时返回完整测试响应。
	return result
	// 结束当前作用域的代码块或结构体定义。
}

// 使用隔离的 MySQL 测试库验证完整业务和旧数据兼容性。
func TestMySQLIntegration(t *testing.T) {
	// 未显式启用集成测试时跳过临时建库操作。
	if os.Getenv("GO_INTEGRATION_TEST") != "1" {
		// 跳过未启用的数据库集成测试，并给出启用方法。
		t.Skip("set GO_INTEGRATION_TEST=1 and ENV_FILE to enable isolated MySQL tests")
		// 结束当前作用域的代码块或结构体定义。
	}
	// 设置 Gin 的测试模式，减少调试输出。
	gin.SetMode(gin.TestMode)
	// 加载环境配置，并保存可能发生的配置错误。
	cfg, err := config.Load()
	// 当前操作失败时进入错误处理分支。
	if err != nil {
		// 记录失败原因并立即结束当前测试。
		t.Fatal(err)
		// 结束当前作用域的代码块或结构体定义。
	}
	// 先不选择业务数据库，以便创建隔离测试库。
	cfg.DBName = ""
	// 创建供测试建库或清理使用的底层 MySQL 连接。
	admin, err := sql.Open("mysql", cfg.DSN())
	// 当前操作失败时进入错误处理分支。
	if err != nil {
		// 记录失败原因并立即结束当前测试。
		t.Fatal(err)
		// 结束当前作用域的代码块或结构体定义。
	}
	// 当前函数退出时关闭该连接或测试服务，释放资源。
	defer admin.Close()
	// 用纳秒时间戳生成本次测试独有的数据库名。
	database := fmt.Sprintf("news_go_test_%d", time.Now().UnixNano())
	// 只允许固定前缀加数字的测试库名称，约束建库和清理范围。
	if !regexp.MustCompile(`^news_go_test_[0-9]+$`).MatchString(database) {
		// 记录失败原因并立即结束当前测试。
		t.Fatal("unsafe test database name")
		// 结束当前作用域的代码块或结构体定义。
	}
	// 创建名称受严格校验的隔离测试数据库。
	if _, err := admin.Exec("CREATE DATABASE `" + database + "` CHARACTER SET utf8mb4"); err != nil {
		// 记录失败原因并立即结束当前测试。
		t.Fatal(err)
		// 结束当前作用域的代码块或结构体定义。
	}
	// 注册测试结束后执行的临时数据库清理函数。
	t.Cleanup(func() {
		// 创建供测试建库或清理使用的底层 MySQL 连接。
		cleanup, err := sql.Open("mysql", cfg.DSN())
		// 当前操作失败时进入错误处理分支。
		if err != nil {
			// 记录失败原因，但允许测试继续完成清理。
			t.Error(err)
			// 结束当前处理，错误或响应已在前面的语句中处理。
			return
			// 结束当前作用域的代码块或结构体定义。
		}
		// 当前函数退出时关闭该连接或测试服务，释放资源。
		defer cleanup.Close()
		// 仅删除本次测试创建的数据库。
		if _, err := cleanup.Exec("DROP DATABASE `" + database + "`"); err != nil {
			// 记录失败原因，但允许测试继续完成清理。
			t.Error(err)
			// 结束当前作用域的代码块或结构体定义。
		}
		// 结束事务、清理函数或子测试的回调调用。
	})
	// 后续查询只连接刚创建的隔离测试库。
	cfg.DBName = database
	// 连接 MySQL 并关闭 SQL 日志，避免输出敏感字段。
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	// 当前操作失败时进入错误处理分支。
	if err != nil {
		// 记录失败原因并立即结束当前测试。
		t.Fatal(err)
		// 结束当前作用域的代码块或结构体定义。
	}
	// 取出已成功初始化的数据库连接池供测试关闭。
	sqlDB, _ := db.DB()
	// 当前函数退出时关闭该连接或测试服务，释放资源。
	defer sqlDB.Close()
	// 读取原项目 SQL 文件，为测试复用真实表结构。
	ddl, err := os.ReadFile("../../../docs/02-数据库sql文件/database.sql")
	// 当前操作失败时进入错误处理分支。
	if err != nil {
		// 记录失败原因并立即结束当前测试。
		t.Fatal(err)
		// 结束当前作用域的代码块或结构体定义。
	}
	// 只提取建表语句，忽略原文件的建库、USE 和样例数据。
	statements := regexp.MustCompile("(?s)CREATE TABLE IF NOT EXISTS `.*?;").FindAllString(string(ddl), -1)
	// 确认从原 SQL 文件提取到预期的八张表结构。
	if len(statements) != 8 {
		// 记录失败原因并立即结束当前测试。
		t.Fatal("unexpected source database schema")
		// 结束当前作用域的代码块或结构体定义。
	}
	// 逐条执行从原始 SQL 提取的建表语句。
	for _, statement := range statements {
		// 在隔离测试库中执行一条原始建表语句。
		if err := db.Exec(statement).Error; err != nil {
			// 记录失败原因并立即结束当前测试。
			t.Fatal(err)
			// 结束当前作用域的代码块或结构体定义。
		}
		// 结束当前作用域的代码块或结构体定义。
	}
	// 构造测试数据模型，供隔离数据库验证。
	category := model.Category{ID: 1, Name: "头条", SortOrder: 1}
	// 写入当前模型记录，并保存数据库生成的主键。
	if err := db.Create(&category).Error; err != nil {
		// 记录失败原因并立即结束当前测试。
		t.Fatal(err)
		// 结束当前作用域的代码块或结构体定义。
	}
	// 创建 ID 与历史记录 ID 不同的测试新闻。
	for _, id := range []uint64{11, 12} {
		// 构造测试数据模型，供隔离数据库验证。
		n := model.News{ID: id, Title: fmt.Sprintf("测试新闻%d", id), CategoryID: 1, Content: "测试正文", PublishTime: time.Now().In(cfg.Location)}
		// 写入当前模型记录，并保存数据库生成的主键。
		if err := db.Create(&n).Error; err != nil {
			// 记录失败原因并立即结束当前测试。
			t.Fatal(err)
			// 结束当前作用域的代码块或结构体定义。
		}
		// 结束当前作用域的代码块或结构体定义。
	}
	// 启动测试专用的内存 Redis，并注册自动清理。
	r := miniredis.RunT(t)
	// 创建配置好的 Redis 客户端。
	redisClient := redis.NewClient(&redis.Options{Addr: r.Addr(), MaxRetries: -1, ContextTimeoutEnabled: true})
	// 当前函数退出时关闭该连接或测试服务，释放资源。
	defer redisClient.Close()
	// 用虚构测试密钥替代真实 AI 密钥。
	cfg.AIKey = "server-only-test-key"
	// 组合数据库、缓存与配置，创建业务服务实例。
	app := New(db, &cache.Store{Client: redisClient, Prefix: "integration:"}, cfg)
	// 创建携带对应用户令牌的测试请求辅助器。
	api := testAPI{t: t, router: app.Router()}
	// 定义仅用于隔离测试库的用户名和密码。
	creds := map[string]string{"username": "go-test-user", "password": "test-password"}

	// 运行注册、登录与令牌替换子测试。
	t.Run("registration_login_and_token_rotation", func(t *testing.T) {
		// 让请求辅助器把失败记录到当前子测试。
		api.t = t
		// 调用 POST /api/user/register，并断言返回状态码 200。
		data := api.request("POST", "/api/user/register", creds, 200)
		// 更新测试请求使用的令牌，验证相应用户或登录状态。
		api.token = data["token"].(string)
		// 确认注册响应没有泄露数据库中的密码字段。
		if strings.Contains(fmt.Sprint(data), "password") {
			// 记录失败原因并立即结束当前测试。
			t.Fatal("password exposed")
			// 结束当前作用域的代码块或结构体定义。
		}
		// 调用 POST /api/user/register，并断言返回状态码 400。
		api.request("POST", "/api/user/register", creds, 400)
		// 调用 POST /api/user/login，并断言返回状态码 401。
		api.request("POST", "/api/user/login", map[string]string{"username": "go-test-user", "password": "wrong"}, 401)
		// 调用 POST /api/user/login，并断言返回状态码 200。
		newToken := api.request("POST", "/api/user/login", creds, 200)["token"].(string)
		// 调用 GET /api/user/info，并断言返回状态码 401。
		api.request("GET", "/api/user/info", nil, 401)
		// 更新测试请求使用的令牌，验证相应用户或登录状态。
		api.token = "Bearer " + newToken
		// 调用 GET /api/user/info，并断言返回状态码 200。
		info := api.request("GET", "/api/user/info", nil, 200)
		// 确认令牌返回的资料属于刚登录的用户。
		if info["username"] != creds["username"] {
			// 记录失败原因并立即结束当前测试。
			t.Fatal("incorrect user")
			// 结束当前作用域的代码块或结构体定义。
		}
		// 结束事务、清理函数或子测试的回调调用。
	})
	// 运行旧密码哈希和令牌兼容子测试。
	t.Run("python_bcrypt_and_existing_token_compatibility", func(t *testing.T) {
		// 此$2b$哈希由项目原有passlib代码生成。
		// 构造测试数据模型，供隔离数据库验证。
		u := model.User{Username: "legacy-user", Password: "$2b$12$.Nujxqtcickr7smb1OWO/.aZlZTSZ1z2O4KnRJ8AXmJXDlOzM6ZIG"}
		// 写入当前模型记录，并保存数据库生成的主键。
		if err := db.Create(&u).Error; err != nil {
			// 记录失败原因并立即结束当前测试。
			t.Fatal(err)
			// 结束当前作用域的代码块或结构体定义。
		}
		// 构造测试数据模型，供隔离数据库验证。
		old := model.UserToken{UserID: u.ID, Token: "python-existing-token", ExpiresAt: time.Now().In(cfg.Location).Add(time.Hour)}
		// 写入当前模型记录，并保存数据库生成的主键。
		if err := db.Create(&old).Error; err != nil {
			// 记录失败原因并立即结束当前测试。
			t.Fatal(err)
			// 结束当前作用域的代码块或结构体定义。
		}
		// 创建携带对应用户令牌的测试请求辅助器。
		legacyAPI := testAPI{t: t, router: api.router, token: old.Token}
		// 调用 GET /api/user/info，并断言返回状态码 200。
		legacyAPI.request("GET", "/api/user/info", nil, 200)
		// 调用 POST /api/user/login，并断言返回状态码 200。
		legacyAPI.request("POST", "/api/user/login", map[string]string{"username": u.Username, "password": "legacy-password"}, 200)
		// 调用 GET /api/user/info，并断言返回状态码 401。
		legacyAPI.request("GET", "/api/user/info", nil, 401)
		// 构造测试数据模型，供隔离数据库验证。
		expired := model.UserToken{UserID: u.ID, Token: "expired-python-token", ExpiresAt: time.Now().In(cfg.Location).Add(-time.Hour)}
		// 写入当前模型记录，并保存数据库生成的主键。
		if err := db.Create(&expired).Error; err != nil {
			// 记录失败原因并立即结束当前测试。
			t.Fatal(err)
			// 结束当前作用域的代码块或结构体定义。
		}
		// 更新测试请求使用的令牌，验证相应用户或登录状态。
		legacyAPI.token = expired.Token
		// 调用 GET /api/user/info，并断言返回状态码 401。
		legacyAPI.request("GET", "/api/user/info", nil, 401)
		// 结束事务、清理函数或子测试的回调调用。
	})
	// 运行资料和密码修改子测试。
	t.Run("profile_and_password", func(t *testing.T) {
		// 让请求辅助器把失败记录到当前子测试。
		api.t = t
		// 调用 PUT /api/user/update，并断言返回状态码 200。
		profile := api.request("PUT", "/api/user/update", map[string]string{"nickname": "昵称", "bio": "简介", "gender": "female"}, 200)
		// 确认资料更新确实持久化了请求中的昵称。
		if profile["nickname"] != "昵称" {
			// 记录失败原因并立即结束当前测试。
			t.Fatal("profile not updated")
			// 结束当前作用域的代码块或结构体定义。
		}
		// 调用 PUT /api/user/update，并断言返回状态码 200。
		api.request("PUT", "/api/user/update", map[string]string{}, 200)
		// 调用 PUT /api/user/update，并断言返回状态码 422。
		api.request("PUT", "/api/user/update", map[string]string{"gender": "invalid"}, 422)
		// 调用 PUT /api/user/password，并断言返回状态码 400。
		api.request("PUT", "/api/user/password", map[string]string{"oldPassword": "wrong", "newPassword": "new-password"}, 400)
		// 调用 PUT /api/user/password，并断言返回状态码 200。
		api.request("PUT", "/api/user/password", map[string]string{"oldPassword": "test-password", "newPassword": "new-password"}, 200)
		// 调用 POST /api/user/login，并断言返回状态码 401。
		api.request("POST", "/api/user/login", creds, 401)
		// 调用 POST /api/user/login，并断言返回状态码 200。
		data := api.request("POST", "/api/user/login", map[string]string{"username": "go-test-user", "password": "new-password"}, 200)
		// 更新测试请求使用的令牌，验证相应用户或登录状态。
		api.token = data["token"].(string)
		// 结束事务、清理函数或子测试的回调调用。
	})
	// 运行新闻分页、缓存和浏览量子测试。
	t.Run("news_pagination_cache_and_views", func(t *testing.T) {
		// 让请求辅助器把失败记录到当前子测试。
		api.t = t
		// 调用 GET /api/news/categories，并断言返回状态码 200。
		api.request("GET", "/api/news/categories", nil, 200)
		// 调用 GET /api/news/list?categoryId=1&pageSize=1，并断言返回状态码 200。
		first := api.request("GET", "/api/news/list?categoryId=1&pageSize=1", nil, 200)
		// 确认第一页返回两条总记录，并表示仍有下一页。
		if first["hasMore"] != true || first["total"] != float64(2) {
			// 记录失败原因并立即结束当前测试。
			t.Fatal("incorrect pagination")
			// 结束当前作用域的代码块或结构体定义。
		}
		// 取出列表第一条记录并转换为可检查字段的对象。
		item := first["list"].([]any)[0].(map[string]any)
		// 确认发布时间和分类字段采用前端要求的命名与值。
		if item["publishTime"] == nil || item["categoryId"] != float64(1) {
			// 记录失败原因并立即结束当前测试。
			t.Fatal("frontend fields missing")
			// 结束当前作用域的代码块或结构体定义。
		}
		// 调用 GET /api/news/list?categoryId=1&pageSize=1，并断言返回状态码 200。
		api.request("GET", "/api/news/list?categoryId=1&pageSize=1", nil, 200)
		// 确认新闻列表查询确实写入了对应分类和页码的 Redis 缓存。
		if !r.Exists("integration:list:1:1:1") {
			// 记录失败原因并立即结束当前测试。
			t.Fatal("Redis cache not populated")
			// 结束当前作用域的代码块或结构体定义。
		}
		// 调用 GET /api/news/list?categoryId=1&pageSize=1&page=2，并断言返回状态码 200。
		second := api.request("GET", "/api/news/list?categoryId=1&pageSize=1&page=2", nil, 200)
		// 确认最后一页不会误报仍有更多数据。
		if second["hasMore"] != false {
			// 记录失败原因并立即结束当前测试。
			t.Fatal("incorrect last page")
			// 结束当前作用域的代码块或结构体定义。
		}
		// 调用 GET /api/news/detail?id=11，并断言返回状态码 200。
		detail := api.request("GET", "/api/news/detail?id=11", nil, 200)
		// 验证相关新闻按当前分类查询，并排除当前新闻。
		related := detail["relatedNews"].([]any)
		// 当前测试分类只有两篇新闻，相关新闻应只包含 ID 为 12 的另一篇。
		if len(related) != 1 || related[0].(map[string]any)["id"] != float64(12) {
			// 查询参数顺序错误或漏掉相关新闻时立即报告失败。
			t.Fatal("related news must match the category and exclude the current item")
			// 结束断言分支。
		}
		// 调用 GET /api/news/detail?id=11，并断言返回状态码 200。
		cached := api.request("GET", "/api/news/detail?id=11", nil, 200)
		// 确认命中详情缓存后，数据库浏览量仍会增加一。
		if cached["views"].(float64) != detail["views"].(float64)+1 {
			// 记录失败原因并立即结束当前测试。
			t.Fatal("cached views did not increase")
			// 结束当前作用域的代码块或结构体定义。
		}
		// 调用 GET /api/news/detail?id=999，并断言返回状态码 404。
		api.request("GET", "/api/news/detail?id=999", nil, 404)
		// 关闭内存 Redis，模拟缓存服务不可用。
		r.Close()
		// 调用 GET /api/news/list?categoryId=1，并断言返回状态码 200。
		api.request("GET", "/api/news/list?categoryId=1", nil, 200)
		// 结束事务、清理函数或子测试的回调调用。
	})
	// 运行收藏完整流程子测试。
	t.Run("favorites", func(t *testing.T) {
		// 让请求辅助器把失败记录到当前子测试。
		api.t = t
		// 调用 POST /api/favorite/add，并断言返回状态码 200。
		api.request("POST", "/api/favorite/add", map[string]int{"newsId": 11}, 200)
		// 调用 POST /api/favorite/add，并断言返回状态码 400。
		api.request("POST", "/api/favorite/add", map[string]int{"newsId": 11}, 400)
		// 调用 GET /api/favorite/check?newsId=11，并断言返回状态码 200。
		if api.request("GET", "/api/favorite/check?newsId=11", nil, 200)["isFavorite"] != true {
			// 记录失败原因并立即结束当前测试。
			t.Fatal("favorite missing")
			// 结束当前作用域的代码块或结构体定义。
		}
		// 调用 GET /api/favorite/list，并断言返回状态码 200。
		list := api.request("GET", "/api/favorite/list", nil, 200)
		// 取出列表第一条记录并转换为可检查字段的对象。
		item := list["list"].([]any)[0].(map[string]any)
		// 确认收藏列表包含收藏记录 ID、收藏时间和新闻发布时间。
		if item["favoriteId"] == nil || item["favoriteTime"] == nil || item["publishTime"] == nil {
			// 记录失败原因并立即结束当前测试。
			t.Fatal("favorite fields missing")
			// 结束当前作用域的代码块或结构体定义。
		}
		// 调用 DELETE /api/favorite/remove?newsId=11，并断言返回状态码 200。
		api.request("DELETE", "/api/favorite/remove?newsId=11", nil, 200)
		// 调用 DELETE /api/favorite/remove?newsId=11，并断言返回状态码 404。
		api.request("DELETE", "/api/favorite/remove?newsId=11", nil, 404)
		// 调用 POST /api/favorite/add，并断言返回状态码 200。
		api.request("POST", "/api/favorite/add", map[string]int{"newsId": 12}, 200)
		// 调用 DELETE /api/favorite/clear，并断言返回状态码 200。
		api.request("DELETE", "/api/favorite/clear", nil, 200)
		// 调用 GET /api/favorite/list，并断言返回状态码 200。
		if api.request("GET", "/api/favorite/list", nil, 200)["total"] != float64(0) {
			// 记录失败原因并立即结束当前测试。
			t.Fatal("favorites not cleared")
			// 结束当前作用域的代码块或结构体定义。
		}
		// 结束事务、清理函数或子测试的回调调用。
	})
	// 运行历史 ID 和用户数据隔离子测试。
	t.Run("history_record_ids_and_user_isolation", func(t *testing.T) {
		// 让请求辅助器把失败记录到当前子测试。
		api.t = t
		// 调用 POST /api/history/add，并断言返回状态码 200。
		first := api.request("POST", "/api/history/add", map[string]int{"newsId": 11}, 200)
		// 调用 POST /api/history/add，并断言返回状态码 200。
		repeat := api.request("POST", "/api/history/add", map[string]int{"newsId": 11}, 200)
		// 确认重复浏览相同新闻会更新旧记录，而不是创建重复历史。
		if first["id"] != repeat["id"] {
			// 记录失败原因并立即结束当前测试。
			t.Fatal("duplicate history created")
			// 结束当前作用域的代码块或结构体定义。
		}
		// 调用 POST /api/history/add，并断言返回状态码 200。
		api.request("POST", "/api/history/add", map[string]int{"newsId": 12}, 200)
		// 调用 GET /api/history/list，并断言返回状态码 200。
		list := api.request("GET", "/api/history/list", nil, 200)
		// 确认浏览两篇新闻后恰好产生两条历史记录。
		if list["total"] != float64(2) {
			// 记录失败原因并立即结束当前测试。
			t.Fatal("incorrect history count")
			// 结束当前作用域的代码块或结构体定义。
		}
		// 取出列表第一条记录并转换为可检查字段的对象。
		item := list["list"].([]any)[0].(map[string]any)
		// 确认历史列表包含记录 ID、浏览时间和新闻发布时间。
		if item["historyId"] == nil || item["viewTime"] == nil || item["publishTime"] == nil {
			// 记录失败原因并立即结束当前测试。
			t.Fatal("history fields missing")
			// 结束当前作用域的代码块或结构体定义。
		}
		// 创建携带对应用户令牌的测试请求辅助器。
		other := testAPI{t: t, router: api.router}
		// 调用 POST /api/user/register，并断言返回状态码 200。
		other.token = other.request("POST", "/api/user/register", map[string]string{"username": "other-user", "password": "test-password"}, 200)["token"].(string)
		// 使用历史记录主键构造删除路径，而非使用新闻 ID。
		path := fmt.Sprintf("/api/history/delete/%.0f", first["id"].(float64))
		// 用其他用户令牌尝试删除记录，断言不能删除别人的数据。
		other.request("DELETE", path, nil, 404)
		// 用记录所属用户的令牌删除该历史记录，断言成功。
		api.request("DELETE", path, nil, 200)
		// 调用 GET /api/history/list，并断言返回状态码 200。
		if api.request("GET", "/api/history/list", nil, 200)["total"] != float64(1) {
			// 记录失败原因并立即结束当前测试。
			t.Fatal("wrong history deleted")
			// 结束当前作用域的代码块或结构体定义。
		}
		// 调用 DELETE /api/history/clear，并断言返回状态码 200。
		api.request("DELETE", "/api/history/clear", nil, 200)
		// 结束事务、清理函数或子测试的回调调用。
	})
	// 运行AI 认证、流响应和错误脱敏子测试。
	t.Run("ai_authenticated_stream_and_sanitized_errors", func(t *testing.T) {
		// 让请求辅助器把失败记录到当前子测试。
		api.t = t
		// 启动本地模拟 AI 服务，避免调用真实模型。
		upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// 确认上游收到的是后端密钥，而不是浏览器登录令牌。
			if req.Header.Get("Authorization") != "Bearer server-only-test-key" {
				// 记录失败原因，但允许测试继续完成清理。
				t.Error("wrong server credentials")
				// 结束当前作用域的代码块或结构体定义。
			}
			// 让模拟 AI 上游声明返回 SSE 数据流。
			w.Header().Set("Content-Type", "text/event-stream")
			// 写入模拟 AI 上游的回复或错误文本。
			_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"hello\"}}]}\n\ndata: [DONE]\n\n"))
			// 结束模拟 HTTP 处理器并完成测试服务创建。
		}))
		// 当前函数退出时关闭该连接或测试服务，释放资源。
		defer upstream.Close()
		// 把 AI 请求连接到本地模拟服务，避免访问真实上游。
		app.Service.Config.AIEndpoint = upstream.URL
		// 把 AI 请求连接到本地模拟服务，避免访问真实上游。
		app.Service.HTTP = upstream.Client()
		// 定义测试 AI 消息请求体。
		body := `{"messages":[{"role":"user","content":"hello"}]}`
		// 构造发往被测接口的 HTTP 请求。
		req := httptest.NewRequest("POST", "/api/ai/chat", strings.NewReader(body))
		// 设置测试请求的认证头或 JSON 内容类型。
		req.Header.Set("Authorization", api.token)
		// 设置测试请求的认证头或 JSON 内容类型。
		req.Header.Set("Content-Type", "application/json")
		// 创建测试响应记录器，接收状态码、响应头和响应体。
		w := httptest.NewRecorder()
		// 通过真实 Gin 路由执行测试请求。
		api.router.ServeHTTP(w, req)
		// 确认流式回复成功、包含结束标记，且没有把后端密钥返回前端。
		if w.Code != 200 || !strings.Contains(w.Body.String(), "[DONE]") || strings.Contains(w.Body.String(), app.Service.Config.AIKey) {
			// 记录失败原因并立即结束当前测试。
			t.Fatal("invalid AI stream")
			// 结束当前作用域的代码块或结构体定义。
		}
		// 清空测试配置中的密钥，验证未配置时的处理。
		app.Service.Config.AIKey = ""
		// 调用 POST /api/ai/chat，并断言返回状态码 503。
		api.request("POST", "/api/ai/chat", map[string]any{"messages": []dto.ChatMessage{{Role: "user", Content: "hello"}}}, 503)
		// 恢复虚构测试密钥，验证上游失败时的处理。
		app.Service.Config.AIKey = "server-only-test-key"
		// 启动本地模拟 AI 服务，避免调用真实模型。
		failed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// 模拟上游返回认证失败状态。
			w.WriteHeader(401)
			// 写入模拟 AI 上游的回复或错误文本。
			_, _ = w.Write([]byte(app.Service.Config.AIKey))
			// 结束模拟 HTTP 处理器并完成测试服务创建。
		}))
		// 当前函数退出时关闭该连接或测试服务，释放资源。
		defer failed.Close()
		// 把 AI 请求连接到本地模拟服务，避免访问真实上游。
		app.Service.Config.AIEndpoint = failed.URL
		// 调用 POST /api/ai/chat，并断言返回状态码 502。
		api.request("POST", "/api/ai/chat", map[string]any{"messages": []dto.ChatMessage{{Role: "user", Content: "hello"}}}, 502)
		// 结束事务、清理函数或子测试的回调调用。
	})
	// 结束当前作用域的代码块或结构体定义。
}
