// 声明独立业务逻辑包。
package service

// 导入本文件使用的标准库、第三方依赖和内部包。
import (
	// 引入context，用于请求取消与超时上下文。
	"context"
	// 引入errors，用于错误类型匹配和构造。
	"errors"
	// 引入fmt，用于缓存键和提示消息格式化。
	"fmt"
	// 引入time，用于时间和有效期计算。
	"time"

	// 引入github.com/prejudicing/news_py_go/backend_go/internal/dto，用于接口参数和响应结构。
	"github.com/prejudicing/news_py_go/backend_go/internal/dto"
	// 引入github.com/prejudicing/news_py_go/backend_go/internal/model，用于数据库实体。
	"github.com/prejudicing/news_py_go/backend_go/internal/model"
	// 引入github.com/prejudicing/news_py_go/backend_go/internal/repository，用于仓储接口和统一错误。
	"github.com/prejudicing/news_py_go/backend_go/internal/repository"
	// 结束当前声明或代码作用域。
)

// 读取排序后的新闻分类。
func (s *Service) Categories(ctx context.Context, skip, limit int) ([]model.Category, error) {
	// 构造或保存包含查询参数的缓存键。
	key := fmt.Sprintf("categories:%d:%d", skip, limit)
	// 创建非空列表容器，使无数据时 JSON 返回空数组。
	items := make([]model.Category, 0)
	// 缓存未命中或 Redis 不可用时从仓储读取数据。
	if !s.Cache.Get(ctx, key, &items) {
		// 声明操作错误。
		var err error
		// 读取排序后的新闻分类并保存操作结果。
		items, err = s.Repo.Categories(ctx, skip, limit)
		// 检查本次操作是否失败，必要时提前返回错误。
		if err != nil {
			// 返回当前结果与错误，由调用方决定响应或事务回滚。
			return nil, err
			// 结束当前声明或代码作用域。
		}
		// 写入带有效期的缓存，缓存失败时业务仍可继续。
		s.Cache.Set(ctx, key, items, 2*time.Hour)
		// 结束当前声明或代码作用域。
	}
	// 返回当前计算或构造的结果。
	return items, nil
	// 结束当前声明或代码作用域。
}

// 读取指定分类的分页新闻与总数。
func (s *Service) NewsList(ctx context.Context, category, page, size int) (dto.Page[dto.NewsItem], error) {
	// 统计指定分类的新闻总数并保存操作结果。
	total, err := s.Repo.CountNews(ctx, category)
	// 检查本次操作是否失败，必要时提前返回错误。
	if err != nil {
		// 返回当前结果与错误，由调用方决定响应或事务回滚。
		return dto.Page[dto.NewsItem]{}, err
		// 结束当前声明或代码作用域。
	}
	// 创建非空列表容器，使无数据时 JSON 返回空数组。
	items := make([]dto.NewsItem, 0)
	// 构造或保存包含查询参数的缓存键。
	key := fmt.Sprintf("list:%d:%d:%d", category, page, size)
	// 缓存未命中或 Redis 不可用时从仓储读取数据。
	if !s.Cache.Get(ctx, key, &items) {
		// 按偏移和数量查询指定分类的新闻并保存操作结果。
		rows, err := s.Repo.ListNews(ctx, category, (page-1)*size, size)
		// 检查本次操作是否失败，必要时提前返回错误。
		if err != nil {
			// 返回当前结果与错误，由调用方决定响应或事务回滚。
			return dto.Page[dto.NewsItem]{}, err
			// 结束当前声明或代码作用域。
		}
		// 遍历仓储查询结果，转换为前端响应结构。
		for _, row := range rows {
			// 把当前实体的响应字段加入列表。
			items = append(items, newsDTO(row))
			// 结束当前声明或代码作用域。
		}
		// 写入带有效期的缓存，缓存失败时业务仍可继续。
		s.Cache.Set(ctx, key, items, 30*time.Minute)
		// 结束当前声明或代码作用域。
	}
	// 构造接口响应，返回数据和成功标记。
	return dto.Page[dto.NewsItem]{List: items, Total: total, HasMore: int64((page-1)*size+len(items)) < total}, nil
	// 结束当前声明或代码作用域。
}

// 读取新闻详情、累计浏览量并查询相关新闻。
func (s *Service) NewsDetail(ctx context.Context, id uint64) (dto.NewsDetail, error) {
	// 声明当前实体。
	var item model.News
	// 构造或保存包含查询参数的缓存键。
	key := fmt.Sprintf("detail:%d", id)
	// 缓存未命中或 Redis 不可用时从仓储读取数据。
	if !s.Cache.Get(ctx, key, &item) {
		// 声明操作错误。
		var err error
		// 按主键查找新闻实体并保存操作结果。
		item, err = s.Repo.NewsByID(ctx, id)
		// 检查错误是否表示记录不存在，选择相应的业务分支。
		if errors.Is(err, repository.ErrNotFound) {
			// 返回校验失败对应的业务状态码和消息。
			return dto.NewsDetail{}, problem(404, "新闻不存在")
			// 结束当前声明或代码作用域。
		}
		// 检查本次操作是否失败，必要时提前返回错误。
		if err != nil {
			// 返回当前结果与错误，由调用方决定响应或事务回滚。
			return dto.NewsDetail{}, err
			// 结束当前声明或代码作用域。
		}
		// 写入带有效期的缓存，缓存失败时业务仍可继续。
		s.Cache.Set(ctx, key, item, 5*time.Minute)
		// 结束当前声明或代码作用域。
	}
	// 在数据库事务中递增并读取最新浏览量并保存操作结果。
	views, err := s.Repo.IncreaseViews(ctx, id)
	// 检查本次操作是否失败，必要时提前返回错误。
	if err != nil {
		// 返回当前结果与错误，由调用方决定响应或事务回滚。
		return dto.NewsDetail{}, err
		// 结束当前声明或代码作用域。
	}
	// 以数据库的最新浏览量替换缓存中的旧值。
	item.Views = views
	// 创建非空列表容器，使无数据时 JSON 返回空数组。
	related := make([]dto.RelatedItem, 0)
	// 构造或保存包含新闻和分类的相关新闻缓存键。
	relatedKey := fmt.Sprintf("related:%d:%d", item.ID, item.CategoryID)
	// 缓存未命中或 Redis 不可用时从仓储读取数据。
	if !s.Cache.Get(ctx, relatedKey, &related) {
		// 查询同分类的其他热门新闻并保存操作结果。
		rows, err := s.Repo.RelatedNews(ctx, item.ID, item.CategoryID)
		// 检查本次操作是否失败，必要时提前返回错误。
		if err != nil {
			// 返回当前结果与错误，由调用方决定响应或事务回滚。
			return dto.NewsDetail{}, err
			// 结束当前声明或代码作用域。
		}
		// 遍历仓储查询结果，转换为前端响应结构。
		for _, row := range rows {
			// 把当前相关新闻摘要加入响应列表。
			related = append(related, dto.RelatedItem{ID: row.ID, Title: row.Title, Image: row.Image, Views: row.Views})
			// 结束当前声明或代码作用域。
		}
		// 写入带有效期的缓存，缓存失败时业务仍可继续。
		s.Cache.Set(ctx, relatedKey, related, 30*time.Minute)
		// 结束当前声明或代码作用域。
	}
	// 构造接口响应，返回数据和成功标记。
	return dto.NewsDetail{ID: item.ID, Title: item.Title, Content: item.Content, Image: item.Image, Author: item.Author, PublishTime: Timestamp(item.PublishTime), CategoryID: item.CategoryID, Views: item.Views, RelatedNews: related}, nil
	// 结束当前声明或代码作用域。
}
