// 声明业务接口与路由包。
package server

// 开始声明本文件使用的标准库、第三方库和项目内部包。
import (
	// 引入 bytes，用于将字节数据包装为可读取的请求体。
	"bytes"
	// 引入 encoding/json，用于编码和解码 JSON 数据。
	"encoding/json"
	// 引入 net/http/httptest，用于在测试中模拟 HTTP 请求、响应和上游服务。
	"net/http/httptest"
	// 引入 testing，用于编写和运行 Go 测试。
	"testing"

	// 引入 github.com/gin-gonic/gin，用于使用 Gin 路由、请求上下文和 JSON 响应。
	"github.com/gin-gonic/gin"
	// 引入项目内部的 config 包，复用项目环境配置。
	"github.com/prejudicing/news_py_go/backend_go/internal/config"
	// 结束依赖导入列表。
)

// 测试不依赖数据库的参数验证、认证和错误格式。
func TestValidationAndAuthWithoutDatabase(t *testing.T) {
	// 设置 Gin 的测试模式，减少调试输出。
	gin.SetMode(gin.TestMode)
	// 创建没有数据库依赖的测试路由。
	r := New(nil, nil, config.Config{}).Router()
	// 定义覆盖正常、参数错误、缺少认证和路径错误的测试表。
	cases := []struct {
		// 定义测试请求的方法、地址和请求体。
		method, path, body string
		// 保存 期望的 HTTP 状态码。
		status int
		// 结束测试用例类型定义，开始填写用例数据。
	}{
		// 定义 GET / 的请求与期望状态码。
		{"GET", "/", "", 200},
		// 定义 OPTIONS /api/news/list 的请求与期望状态码。
		{"OPTIONS", "/api/news/list", "", 204},
		// 定义 GET /api/news/list 的请求与期望状态码。
		{"GET", "/api/news/list", "", 422},
		// 定义 GET /api/news/list?categoryId=1&pageSize=0 的请求与期望状态码。
		{"GET", "/api/news/list?categoryId=1&pageSize=0", "", 422},
		// 定义 GET /api/news/list?categoryId=1&page=-1 的请求与期望状态码。
		{"GET", "/api/news/list?categoryId=1&page=-1", "", 422},
		// 定义 GET /api/news/detail?id=invalid 的请求与期望状态码。
		{"GET", "/api/news/detail?id=invalid", "", 422},
		// 定义 GET /api/news/categories?skip=-1 的请求与期望状态码。
		{"GET", "/api/news/categories?skip=-1", "", 422},
		// 定义 GET /api/favorite/list 的请求与期望状态码。
		{"GET", "/api/favorite/list", "", 401},
		// 定义 POST /api/ai/chat 的请求与期望状态码。
		{"POST", "/api/ai/chat", `{"messages":[]}`, 401},
		// 定义 POST /api/user/register 的请求与期望状态码。
		{"POST", "/api/user/register", `{"username":"test"}`, 422},
		// 定义 POST /api/user/login 的请求与期望状态码。
		{"POST", "/api/user/login", "invalid", 422},
		// 定义 GET /missing 的请求与期望状态码。
		{"GET", "/missing", "", 404},
		// 定义 POST /api/news/list 的请求与期望状态码。
		{"POST", "/api/news/list", "", 405},
		// 结束当前作用域的代码块或结构体定义。
	}
	// 逐个执行表格中定义的参数验证用例。
	for _, tc := range cases {
		// 运行当前表格用例子测试。
		t.Run(tc.method+tc.path, func(t *testing.T) {
			// 构造发往被测接口的 HTTP 请求。
			req := httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(tc.body))
			// 设置测试请求的认证头或 JSON 内容类型。
			req.Header.Set("Content-Type", "application/json")
			// 创建测试响应记录器，接收状态码、响应头和响应体。
			w := httptest.NewRecorder()
			// 通过真实 Gin 路由执行测试请求。
			r.ServeHTTP(w, req)
			// 比较真实响应状态码与该测试用例的期望值。
			if w.Code != tc.status {
				// 记录失败原因并立即结束当前测试。
				t.Fatalf("status %d, want %d: %s", w.Code, tc.status, w.Body.String())
				// 结束当前作用域的代码块或结构体定义。
			}
			// 检查跨域响应头，确保浏览器可以访问接口。
			if w.Header().Get("Access-Control-Allow-Origin") != "*" {
				// 记录失败原因并立即结束当前测试。
				t.Fatal("CORS header missing")
				// 结束当前作用域的代码块或结构体定义。
			}
			// 错误响应还需要验证统一的 JSON 响应格式。
			if tc.status >= 400 {
				// 声明变量，用于接收测试响应的 JSON 字段。
				var body struct {
					// 保存 业务状态码。JSON 标签指定前端字段名。
					Code int `json:"code"`
					// 保存 结果提示。JSON 标签指定前端字段名。
					Message string `json:"message"`
					// 保存 响应数据。JSON 标签指定前端字段名。
					Data any `json:"data"`
					// 结束当前作用域的代码块或结构体定义。
				}
				// 把 JSON 数据解码到目标对象，并检查解码错误。
				if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
					// 记录失败原因并立即结束当前测试。
					t.Fatal(err)
					// 结束当前作用域的代码块或结构体定义。
				}
				// 确认错误码与 HTTP 状态一致，提示非空且 data 为 null。
				if body.Code != tc.status || body.Message == "" || body.Data != nil {
					// 记录失败原因并立即结束当前测试。
					t.Fatal("invalid error envelope")
					// 结束当前作用域的代码块或结构体定义。
				}
				// 结束当前作用域的代码块或结构体定义。
			}
			// 结束事务、清理函数或子测试的回调调用。
		})
		// 结束当前作用域的代码块或结构体定义。
	}
	// 结束当前作用域的代码块或结构体定义。
}
