// 声明 Redis 缓存包。
package cache

// 开始声明本文件使用的标准库、第三方库和项目内部包。
import (
	// 引入 context，用于管理请求取消、超时和资源释放。
	"context"
	// 引入 encoding/json，用于编码和解码 JSON 数据。
	"encoding/json"
	// 引入 time，用于处理时间、有效期和超时。
	"time"

	// 引入 github.com/redis/go-redis/v9，用于使用 Redis 客户端读取和写入缓存。
	"github.com/redis/go-redis/v9"
	// 结束依赖导入列表。
)

// 封装带命名空间的 Redis JSON 缓存。
type Store struct {
	// 保存 Redis 客户端。
	Client *redis.Client
	// 保存 缓存键前缀。
	Prefix string
	// 结束当前作用域的代码块或结构体定义。
}

// 读取并解码 Redis 缓存，失败时按未命中处理。
func (s *Store) Get(ctx context.Context, key string, value any) bool {
	// 没有配置 Redis 客户端时直接使用缓存降级。
	if s == nil || s.Client == nil {
		// 返回失败或未命中的标记。
		return false
		// 结束当前作用域的代码块或结构体定义。
	}
	// 将本次缓存操作限制在半秒内。
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	// 函数退出时取消上下文，释放计时器资源。
	defer cancel()
	// 读取带 Go 命名空间前缀的缓存值。
	data, err := s.Client.Get(ctx, s.Prefix+key).Bytes()
	// 只有缓存读取成功且 JSON 解码成功时才认为命中。
	return err == nil && json.Unmarshal(data, value) == nil
	// 结束当前作用域的代码块或结构体定义。
}

// 编码并写入带有效期的缓存，失败时保留数据库查询能力。
func (s *Store) Set(ctx context.Context, key string, value any, ttl time.Duration) {
	// 没有配置 Redis 客户端时直接使用缓存降级。
	if s == nil || s.Client == nil {
		// 结束当前处理，错误或响应已在前面的语句中处理。
		return
		// 结束当前作用域的代码块或结构体定义。
	}
	// 把待缓存对象序列化为 JSON 字节。
	data, err := json.Marshal(value)
	// 当前操作失败时进入错误处理分支。
	if err != nil {
		// 结束当前处理，错误或响应已在前面的语句中处理。
		return
		// 结束当前作用域的代码块或结构体定义。
	}
	// 将本次缓存操作限制在半秒内。
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	// 函数退出时取消上下文，释放计时器资源。
	defer cancel()
	// 写入带有效期的缓存，忽略失败以保留数据库回退能力。
	_ = s.Client.Set(ctx, s.Prefix+key, data, ttl).Err()
	// 结束当前作用域的代码块或结构体定义。
}
