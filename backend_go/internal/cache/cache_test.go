// 声明 Redis 缓存包。
package cache

// 开始声明本文件使用的标准库、第三方库和项目内部包。
import (
	// 引入 context，用于管理请求取消、超时和资源释放。
	"context"
	// 引入 testing，用于编写和运行 Go 测试。
	"testing"
	// 引入 time，用于处理时间、有效期和超时。
	"time"

	// 引入 github.com/alicebob/miniredis/v2，用于启动内存 Redis 测试服务。
	"github.com/alicebob/miniredis/v2"
	// 引入 github.com/redis/go-redis/v9，用于使用 Redis 客户端读取和写入缓存。
	"github.com/redis/go-redis/v9"
	// 结束依赖导入列表。
)

// 测试缓存读写、过期、损坏数据与 Redis 故障降级。
func TestCacheRoundTripExpirationAndFallback(t *testing.T) {
	// 启动测试专用的内存 Redis，并注册自动清理。
	r := miniredis.RunT(t)
	// 创建配置好的 Redis 客户端。
	client := redis.NewClient(&redis.Options{Addr: r.Addr(), MaxRetries: -1, ContextTimeoutEnabled: true})
	// 当前函数退出时关闭该连接或测试服务，释放资源。
	defer client.Close()
	// 给测试缓存设置独立命名空间。
	s := &Store{Client: client, Prefix: "go:test:"}
	// 创建测试操作使用的基础上下文。
	ctx := context.Background()
	// 写入测试缓存及其过期时间。
	s.Set(ctx, "news", map[string]int{"id": 7}, time.Minute)
	// 声明用于接收缓存解码结果的字典。
	var value map[string]int
	// 确认缓存能够命中，并正确解码保存的新闻 ID。
	if !s.Get(ctx, "news", &value) || value["id"] != 7 {
		// 记录失败原因并立即结束当前测试。
		t.Fatal("cache round trip failed")
		// 结束当前作用域的代码块或结构体定义。
	}
	// 确认实际 Redis 键包含专用前缀，避免与其他缓存冲突。
	if !r.Exists("go:test:news") {
		// 记录失败原因并立即结束当前测试。
		t.Fatal("namespace missing")
		// 结束当前作用域的代码块或结构体定义。
	}
	// 推进模拟 Redis 时间，让缓存立即过期。
	r.FastForward(2 * time.Minute)
	// 确认过期或服务故障后的缓存不会被当成有效命中。
	if s.Get(ctx, "news", &value) {
		// 记录失败原因并立即结束当前测试。
		t.Fatal("expired value returned")
		// 结束当前作用域的代码块或结构体定义。
	}
	// 写入损坏的 JSON 缓存，验证解码失败时降级。
	r.Set("go:test:bad", "invalid JSON")
	// 确认损坏的 JSON 缓存不会被返回给业务接口。
	if s.Get(ctx, "bad", &value) {
		// 记录失败原因并立即结束当前测试。
		t.Fatal("invalid cache returned")
		// 结束当前作用域的代码块或结构体定义。
	}
	// 关闭内存 Redis，模拟缓存服务不可用。
	r.Close()
	// 确认过期或服务故障后的缓存不会被当成有效命中。
	if s.Get(ctx, "news", &value) {
		// 记录失败原因并立即结束当前测试。
		t.Fatal("unavailable Redis must be a cache miss")
		// 结束当前作用域的代码块或结构体定义。
	}
	// 声明空缓存指针，测试未配置 Redis 时的行为。
	var disabled *Store
	// 确认未配置缓存时读取操作安全返回未命中。
	if disabled.Get(ctx, "news", &value) {
		// 记录失败原因并立即结束当前测试。
		t.Fatal("nil cache must be a miss")
		// 结束当前作用域的代码块或结构体定义。
	}
	// 结束当前作用域的代码块或结构体定义。
}
