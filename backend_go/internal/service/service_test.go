package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/prejudicing/news_py_go/backend_go/internal/cache"
	"github.com/prejudicing/news_py_go/backend_go/internal/model"
	"github.com/prejudicing/news_py_go/backend_go/internal/repository"
)

// newsStore 只实现新闻测试所需方法，让业务测试脱离数据库运行。
type newsStore struct {
	repository.NewsRepository
	reads        int
	views        int64
	relatedCalls int
	t            *testing.T
}

// 按主键查找新闻实体。
func (r *newsStore) NewsByID(ctx context.Context, id uint64) (model.News, error) {
	r.reads++
	if id != 11 {
		return model.News{}, repository.ErrNotFound
	}
	return model.News{ID: 11, CategoryID: 2, Title: "测试新闻", Content: "正文", PublishTime: time.Now()}, nil
}

// 模拟仓储递增浏览量，供业务层验证。
func (r *newsStore) IncreaseViews(ctx context.Context, id uint64) (int64, error) {
	if id != 11 {
		r.t.Fatalf("unexpected news id: %d", id)
	}
	r.views++
	return r.views, nil
}

// 查询同分类的其他热门新闻。
func (r *newsStore) RelatedNews(ctx context.Context, id, category uint64) ([]model.News, error) {
	r.relatedCalls++
	if id != 11 || category != 2 {
		r.t.Fatalf("related query reversed: id=%d category=%d", id, category)
	}
	return []model.News{{ID: 12, CategoryID: 2, Title: "相关新闻"}}, nil
}

// 验证详情缓存命中仍累计浏览量且保留正确的相关新闻。
func TestNewsDetailCacheStillCountsEveryView(t *testing.T) {
	redisServer := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { client.Close() })
	repo := &newsStore{t: t}
	app := NewNewsService(repo, &cache.Store{Client: client, Prefix: "test:"})
	for expected := int64(1); expected <= 2; expected++ {
		detail, err := app.NewsDetail(context.Background(), 11)
		if err != nil {
			t.Fatal(err)
		}
		if detail.Views != expected || len(detail.RelatedNews) != 1 || detail.RelatedNews[0].ID != 12 {
			t.Fatalf("invalid detail: %+v", detail)
		}
	}
	if repo.reads != 1 || repo.relatedCalls != 1 {
		t.Fatalf("cache misses: news=%d related=%d", repo.reads, repo.relatedCalls)
	}
}

// historyTx 验证历史更新使用同一个事务仓储，并保持原记录 ID。
type historyTx struct {
	repository.HistoryRepository
	t        *testing.T
	existing bool
	created  bool
	updated  bool
	failure  error
}

// 按用户 ID 查询，按需锁定该用户行。
func (r *historyTx) UserByID(ctx context.Context, id uint64, lock bool) (model.User, error) {
	if id != 7 || !lock {
		r.t.Fatal("history must lock its owner")
	}
	return model.User{ID: id}, nil
}

// 按主键查找新闻实体。
func (r *historyTx) NewsByID(ctx context.Context, id uint64) (model.News, error) {
	return model.News{ID: id}, nil
}

// 查询当前用户对指定新闻的最新历史记录。
func (r *historyTx) HistoryByNews(ctx context.Context, uid, nid uint64) (model.History, error) {
	if uid != 7 || nid != 11 {
		r.t.Fatal("history ownership lost")
	}
	if !r.existing {
		return model.History{}, repository.ErrNotFound
	}
	return model.History{ID: 99, UserID: uid, NewsID: nid}, nil
}

// 写入新的浏览历史记录。
func (r *historyTx) CreateHistory(ctx context.Context, item *model.History) error {
	r.created = true
	item.ID = 100
	return r.failure
}

// 按历史记录主键更新浏览时间。
func (r *historyTx) UpdateHistory(ctx context.Context, id uint64, at time.Time) error {
	r.updated = true
	if id != 99 || at.IsZero() {
		r.t.Fatal("history update lost record id or time")
	}
	return r.failure
}

// 定义记录事务回调结果的模拟仓储。
type transactionalStore struct {
	repository.HistoryRepository
	tx            *historyTx
	callbackError error
}

// 执行事务回调，失败时回滚全部写入。
func (r *transactionalStore) WithHistoryTransaction(ctx context.Context, fn func(repository.HistoryRepository) error) error {
	r.callbackError = fn(r.tx)
	return r.callbackError
}

// 验证新增和重复浏览使用事务并保留历史主键。
func TestAddHistoryUsesTransactionAndPreservesRecordID(t *testing.T) {
	for _, existing := range []bool{false, true} {
		t.Run(map[bool]string{false: "create", true: "update"}[existing], func(t *testing.T) {
			tx := &historyTx{t: t, existing: existing}
			repo := &transactionalStore{tx: tx}
			app := NewHistoryService(repo, nil)
			item, err := app.AddHistory(context.Background(), 7, 11)
			if err != nil {
				t.Fatal(err)
			}
			if existing && (item.ID != 99 || !tx.updated || tx.created) {
				t.Fatal("existing history was duplicated")
			}
			if !existing && (item.ID != 100 || !tx.created || tx.updated) {
				t.Fatal("new history was not created")
			}
		})
	}
}

// 验证写入失败传回事务回调以触发回滚。
func TestAddHistoryWriteErrorReachesTransaction(t *testing.T) {
	failure := errors.New("write failed")
	tx := &historyTx{t: t, existing: true, failure: failure}
	repo := &transactionalStore{tx: tx}
	_, err := NewHistoryService(repo, nil).AddHistory(context.Background(), 7, 11)
	if !errors.Is(err, failure) || !errors.Is(repo.callbackError, failure) {
		t.Fatal("write failure must reach transaction callback for rollback")
	}
}
