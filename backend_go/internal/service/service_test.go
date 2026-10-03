// 声明独立业务逻辑包。
package service

// 导入本文件使用的标准库、第三方依赖和内部包。
import (
	// 引入context，用于请求取消与超时上下文。
	"context"
	// 引入errors，用于错误类型匹配和构造。
	"errors"
	// 引入testing，用于单元测试和断言。
	"testing"
	// 引入time，用于时间和有效期计算。
	"time"

	// 引入github.com/alicebob/miniredis/v2，用于测试用的内存 Redis。
	"github.com/alicebob/miniredis/v2"
	// 引入github.com/redis/go-redis/v9，用于Redis 客户端。
	"github.com/redis/go-redis/v9"

	// 引入github.com/prejudicing/news_py_go/backend_go/internal/cache，用于Redis 缓存封装。
	"github.com/prejudicing/news_py_go/backend_go/internal/cache"
	// 引入github.com/prejudicing/news_py_go/backend_go/internal/config，用于环境配置。
	"github.com/prejudicing/news_py_go/backend_go/internal/config"
	// 引入github.com/prejudicing/news_py_go/backend_go/internal/model，用于数据库实体。
	"github.com/prejudicing/news_py_go/backend_go/internal/model"
	// 引入github.com/prejudicing/news_py_go/backend_go/internal/repository，用于仓储接口和统一错误。
	"github.com/prejudicing/news_py_go/backend_go/internal/repository"
	// 结束当前声明或代码作用域。
)

// newsStore 只实现新闻测试所需方法，让业务测试脱离数据库运行。
type newsStore struct {
	// 嵌入仓储接口，测试仅替换需要使用的方法。
	repository.Store
	// 保存新闻实体读取次数。
	reads int
	// 保存模拟的最新浏览次数。
	views int64
	// 保存相关新闻读取次数。
	relatedCalls int
	// 保存当前测试对象。
	t *testing.T
	// 结束当前声明或代码作用域。
}

// 按主键查找新闻实体。
func (r *newsStore) NewsByID(ctx context.Context, id uint64) (model.News, error) {
	// 累计模拟仓储的实体读取次数。
	r.reads++
	// 断言业务结果、查询参数或事务行为符合预期。
	if id != 11 {
		// 返回当前计算或构造的结果。
		return model.News{}, repository.ErrNotFound
		// 结束当前声明或代码作用域。
	}
	// 返回当前计算或构造的结果。
	return model.News{ID: 11, CategoryID: 2, Title: "测试新闻", Content: "正文", PublishTime: time.Now()}, nil
	// 结束当前声明或代码作用域。
}

// 模拟仓储递增浏览量，供业务层验证。
func (r *newsStore) IncreaseViews(ctx context.Context, id uint64) (int64, error) {
	// 断言业务结果、查询参数或事务行为符合预期。
	if id != 11 {
		// 报告不符合预期的结果并立即结束当前测试。
		r.t.Fatalf("unexpected news id: %d", id)
		// 结束当前声明或代码作用域。
	}
	// 模拟每次浏览时递增浏览次数。
	r.views++
	// 返回当前计算或构造的结果。
	return r.views, nil
	// 结束当前声明或代码作用域。
}

// 查询同分类的其他热门新闻。
func (r *newsStore) RelatedNews(ctx context.Context, id, category uint64) ([]model.News, error) {
	// 累计相关新闻查询次数以验证缓存命中。
	r.relatedCalls++
	// 断言业务结果、查询参数或事务行为符合预期。
	if id != 11 || category != 2 {
		// 报告不符合预期的结果并立即结束当前测试。
		r.t.Fatalf("related query reversed: id=%d category=%d", id, category)
		// 结束当前声明或代码作用域。
	}
	// 返回当前计算或构造的结果。
	return []model.News{{ID: 12, CategoryID: 2, Title: "相关新闻"}}, nil
	// 结束当前声明或代码作用域。
}

// 验证详情缓存命中仍累计浏览量且保留正确的相关新闻。
func TestNewsDetailCacheStillCountsEveryView(t *testing.T) {
	// 启动测试用 Redis，由测试框架负责清理。
	redisServer := miniredis.RunT(t)
	// 创建连接本地测试 Redis 的客户端。
	client := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	// 注册测试结束后的连接清理操作。
	t.Cleanup(func() { client.Close() })
	// 构造或保存供业务层使用的仓储依赖。
	repo := &newsStore{t: t}
	// 构造或保存注入模拟仓储的测试业务服务。
	app := New(repo, &cache.Store{Client: client, Prefix: "test:"}, config.Config{})
	// 遍历多个测试场景并执行相同的业务验证。
	for expected := int64(1); expected <= 2; expected++ {
		// 读取新闻详情、累计浏览量并查询相关新闻并保存操作结果。
		detail, err := app.NewsDetail(context.Background(), 11)
		// 检查本次操作是否失败，必要时提前返回错误。
		if err != nil {
			// 报告不符合预期的结果并立即结束当前测试。
			t.Fatal(err)
			// 结束当前声明或代码作用域。
		}
		// 断言业务结果、查询参数或事务行为符合预期。
		if detail.Views != expected || len(detail.RelatedNews) != 1 || detail.RelatedNews[0].ID != 12 {
			// 报告不符合预期的结果并立即结束当前测试。
			t.Fatalf("invalid detail: %+v", detail)
			// 结束当前声明或代码作用域。
		}
		// 结束当前声明或代码作用域。
	}
	// 断言业务结果、查询参数或事务行为符合预期。
	if repo.reads != 1 || repo.relatedCalls != 1 {
		// 报告不符合预期的结果并立即结束当前测试。
		t.Fatalf("cache misses: news=%d related=%d", repo.reads, repo.relatedCalls)
		// 结束当前声明或代码作用域。
	}
	// 结束当前声明或代码作用域。
}

// historyTx 验证历史更新使用同一个事务仓储，并保持原记录 ID。
type historyTx struct {
	// 嵌入仓储接口，测试仅替换需要使用的方法。
	repository.Store
	// 保存当前测试对象。
	t *testing.T
	// 保存是否已存在浏览历史。
	existing bool
	// 保存是否执行了新增历史。
	created bool
	// 保存是否执行了历史更新。
	updated bool
	// 保存测试注入的写入错误。
	failure error
	// 结束当前声明或代码作用域。
}

// 按用户 ID 查询，按需锁定该用户行。
func (r *historyTx) UserByID(ctx context.Context, id uint64, lock bool) (model.User, error) {
	// 需要串行更新用户数据时增加行级写锁。
	if id != 7 || !lock {
		// 报告不符合预期的结果并立即结束当前测试。
		r.t.Fatal("history must lock its owner")
		// 结束当前声明或代码作用域。
	}
	// 返回当前计算或构造的结果。
	return model.User{ID: id}, nil
	// 结束当前声明或代码作用域。
}

// 按主键查找新闻实体。
func (r *historyTx) NewsByID(ctx context.Context, id uint64) (model.News, error) {
	// 返回当前计算或构造的结果。
	return model.News{ID: id}, nil
	// 结束当前声明或代码作用域。
}

// 查询当前用户对指定新闻的最新历史记录。
func (r *historyTx) HistoryByNews(ctx context.Context, uid, nid uint64) (model.History, error) {
	// 断言业务结果、查询参数或事务行为符合预期。
	if uid != 7 || nid != 11 {
		// 报告不符合预期的结果并立即结束当前测试。
		r.t.Fatal("history ownership lost")
		// 结束当前声明或代码作用域。
	}
	// 断言业务结果、查询参数或事务行为符合预期。
	if !r.existing {
		// 返回当前计算或构造的结果。
		return model.History{}, repository.ErrNotFound
		// 结束当前声明或代码作用域。
	}
	// 返回当前计算或构造的结果。
	return model.History{ID: 99, UserID: uid, NewsID: nid}, nil
	// 结束当前声明或代码作用域。
}

// 写入新的浏览历史记录。
func (r *historyTx) CreateHistory(ctx context.Context, item *model.History) error {
	// 记录测试事务执行了历史新增。
	r.created = true
	// 模拟数据库给新历史分配主键。
	item.ID = 100
	// 返回当前计算或构造的结果。
	return r.failure
	// 结束当前声明或代码作用域。
}

// 按历史记录主键更新浏览时间。
func (r *historyTx) UpdateHistory(ctx context.Context, id uint64, at time.Time) error {
	// 记录测试事务执行了历史更新。
	r.updated = true
	// 断言业务结果、查询参数或事务行为符合预期。
	if id != 99 || at.IsZero() {
		// 报告不符合预期的结果并立即结束当前测试。
		r.t.Fatal("history update lost record id or time")
		// 结束当前声明或代码作用域。
	}
	// 返回当前计算或构造的结果。
	return r.failure
	// 结束当前声明或代码作用域。
}

// 定义记录事务回调结果的模拟仓储。
type transactionalStore struct {
	// 嵌入仓储接口，测试仅替换需要使用的方法。
	repository.Store
	// 保存事务内使用的仓储实例。
	tx *historyTx
	// 保存事务回调返回的错误。
	callbackError error
	// 结束当前声明或代码作用域。
}

// 执行事务回调，失败时回滚全部写入。
func (r *transactionalStore) Transaction(ctx context.Context, fn func(repository.Store) error) error {
	// 使用事务仓储执行回调并保存其错误。
	r.callbackError = fn(r.tx)
	// 返回当前计算或构造的结果。
	return r.callbackError
	// 结束当前声明或代码作用域。
}

// 验证新增和重复浏览使用事务并保留历史主键。
func TestAddHistoryUsesTransactionAndPreservesRecordID(t *testing.T) {
	// 遍历多个测试场景并执行相同的业务验证。
	for _, existing := range []bool{false, true} {
		// 分别验证创建历史和更新历史两种场景。
		t.Run(map[bool]string{false: "create", true: "update"}[existing], func(t *testing.T) {
			// 构造或保存当前测试使用的事务仓储。
			tx := &historyTx{t: t, existing: existing}
			// 构造或保存供业务层使用的仓储依赖。
			repo := &transactionalStore{tx: tx}
			// 构造或保存注入模拟仓储的测试业务服务。
			app := New(repo, nil, config.Config{})
			// 在用户行锁保护下创建或更新浏览记录并保存操作结果。
			item, err := app.AddHistory(context.Background(), 7, 11)
			// 检查本次操作是否失败，必要时提前返回错误。
			if err != nil {
				// 报告不符合预期的结果并立即结束当前测试。
				t.Fatal(err)
				// 结束当前声明或代码作用域。
			}
			// 断言业务结果、查询参数或事务行为符合预期。
			if existing && (item.ID != 99 || !tx.updated || tx.created) {
				// 报告不符合预期的结果并立即结束当前测试。
				t.Fatal("existing history was duplicated")
				// 结束当前声明或代码作用域。
			}
			// 断言业务结果、查询参数或事务行为符合预期。
			if !existing && (item.ID != 100 || !tx.created || tx.updated) {
				// 报告不符合预期的结果并立即结束当前测试。
				t.Fatal("new history was not created")
				// 结束当前声明或代码作用域。
			}
			// 结束当前回调或复合值构造。
		})
		// 结束当前声明或代码作用域。
	}
	// 结束当前声明或代码作用域。
}

// 验证写入失败传回事务回调以触发回滚。
func TestAddHistoryWriteErrorReachesTransaction(t *testing.T) {
	// 创建实例并注入所需依赖并保存操作结果。
	failure := errors.New("write failed")
	// 构造或保存当前测试使用的事务仓储。
	tx := &historyTx{t: t, existing: true, failure: failure}
	// 构造或保存供业务层使用的仓储依赖。
	repo := &transactionalStore{tx: tx}
	// 在用户行锁保护下创建或更新浏览记录并保存操作结果。
	_, err := New(repo, nil, config.Config{}).AddHistory(context.Background(), 7, 11)
	// 检查错误是否表示记录不存在，选择相应的业务分支。
	if !errors.Is(err, failure) || !errors.Is(repo.callbackError, failure) {
		// 报告不符合预期的结果并立即结束当前测试。
		t.Fatal("write failure must reach transaction callback for rollback")
		// 结束当前声明或代码作用域。
	}
	// 结束当前声明或代码作用域。
}
