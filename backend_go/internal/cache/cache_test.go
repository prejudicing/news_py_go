package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// 测试缓存读写、过期、损坏数据与 Redis 故障降级。
func TestCacheRoundTripExpirationAndFallback(t *testing.T) {
	r := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: r.Addr(), MaxRetries: -1, ContextTimeoutEnabled: true})
	defer client.Close()
	s := &Store{Client: client, Prefix: "go:test:"}
	ctx := context.Background()
	s.Set(ctx, "news", map[string]int{"id": 7}, time.Minute)
	var value map[string]int
	if !s.Get(ctx, "news", &value) || value["id"] != 7 {
		t.Fatal("cache round trip failed")
	}
	if !r.Exists("go:test:news") {
		t.Fatal("namespace missing")
	}
	r.FastForward(2 * time.Minute)
	if s.Get(ctx, "news", &value) {
		t.Fatal("expired value returned")
	}
	r.Set("go:test:bad", "invalid JSON")
	if s.Get(ctx, "bad", &value) {
		t.Fatal("invalid cache returned")
	}
	r.Close()
	if s.Get(ctx, "news", &value) {
		t.Fatal("unavailable Redis must be a cache miss")
	}
	var disabled *Store
	if disabled.Get(ctx, "news", &value) {
		t.Fatal("nil cache must be a miss")
	}
}
