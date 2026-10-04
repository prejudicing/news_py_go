package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// 封装带命名空间的 Redis JSON 缓存。
type Store struct {
	Client *redis.Client
	Prefix string
}

// 读取并解码 Redis 缓存，失败时按未命中处理。
func (s *Store) Get(ctx context.Context, key string, value any) bool {
	if s == nil || s.Client == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	data, err := s.Client.Get(ctx, s.Prefix+key).Bytes()
	return err == nil && json.Unmarshal(data, value) == nil
}

// 编码并写入带有效期的缓存，失败时保留数据库查询能力。
func (s *Store) Set(ctx context.Context, key string, value any, ttl time.Duration) {
	if s == nil || s.Client == nil {
		return
	}
	data, err := json.Marshal(value)
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	_ = s.Client.Set(ctx, s.Prefix+key, data, ttl).Err()
}
