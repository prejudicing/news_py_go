// 声明独立业务逻辑包。
package service

// 导入本文件使用的标准库、第三方依赖和内部包。
import (
	// 引入context，用于请求取消与超时上下文。
	"context"

	// 引入github.com/prejudicing/news_py_go/backend_go/internal/dto，用于接口参数和响应结构。
	"github.com/prejudicing/news_py_go/backend_go/internal/dto"
	// 引入github.com/prejudicing/news_py_go/backend_go/internal/model，用于数据库实体。
	"github.com/prejudicing/news_py_go/backend_go/internal/model"
	// 结束当前声明或代码作用域。
)

// 检查当前用户是否收藏指定新闻。
func (s *Service) CheckFavorite(ctx context.Context, uid, nid uint64) (bool, error) {
	// 统计该用户对指定新闻的收藏，把结果或错误返回给调用方。
	return s.Repo.FavoriteExists(ctx, uid, nid)
	// 结束当前声明或代码作用域。
}

// 检查新闻存在后添加当前用户的收藏。
func (s *Service) AddFavorite(ctx context.Context, uid, nid uint64) (model.Favorite, error) {
	// 检查本次操作是否失败，必要时提前返回错误。
	if _, err := s.Repo.NewsByID(ctx, nid); err != nil {
		// 返回当前结果与错误，由调用方决定响应或事务回滚。
		return model.Favorite{}, err
		// 结束当前声明或代码作用域。
	}
	// 获取应用时区中的当前时间并保存操作结果。
	item := model.Favorite{UserID: uid, NewsID: nid, CreatedAt: s.now()}
	// 写入收藏记录并保存操作结果。
	err := s.Repo.CreateFavorite(ctx, &item)
	// 返回当前结果与错误，由调用方决定响应或事务回滚。
	return item, err
	// 结束当前声明或代码作用域。
}

// 删除当前用户的指定收藏并检查删除结果。
func (s *Service) RemoveFavorite(ctx context.Context, uid, nid uint64) error {
	// 限定用户和新闻删除收藏，返回影响行数并保存操作结果。
	count, err := s.Repo.DeleteFavorite(ctx, uid, nid)
	// 检查本次操作是否失败，必要时提前返回错误。
	if err != nil {
		// 返回当前结果与错误，由调用方决定响应或事务回滚。
		return err
		// 结束当前声明或代码作用域。
	}
	// 没有影响任何记录时返回记录不存在错误。
	if count == 0 {
		// 返回校验失败对应的业务状态码和消息。
		return problem(404, "收藏记录不存在")
		// 结束当前声明或代码作用域。
	}
	// 返回空错误，表示操作成功。
	return nil
	// 结束当前声明或代码作用域。
}

// 查询当前用户的收藏分页列表。
func (s *Service) Favorites(ctx context.Context, uid uint64, page, size int) (dto.Page[dto.FavoriteItem], error) {
	// 查询当前用户的收藏分页列表并保存操作结果。
	rows, total, err := s.Repo.Favorites(ctx, uid, (page-1)*size, size)
	// 检查本次操作是否失败，必要时提前返回错误。
	if err != nil {
		// 返回当前结果与错误，由调用方决定响应或事务回滚。
		return dto.Page[dto.FavoriteItem]{}, err
		// 结束当前声明或代码作用域。
	}
	// 创建非空列表容器，使无数据时 JSON 返回空数组。
	items := make([]dto.FavoriteItem, 0, len(rows))
	// 遍历仓储查询结果，转换为前端响应结构。
	for _, row := range rows {
		// 把当前实体的响应字段加入列表。
		items = append(items, dto.FavoriteItem{NewsItem: newsDTO(row.News), FavoriteTime: Timestamp(row.FavoriteTime), FavoriteID: row.FavoriteID})
		// 结束当前声明或代码作用域。
	}
	// 构造接口响应，返回数据和成功标记。
	return dto.Page[dto.FavoriteItem]{List: items, Total: total, HasMore: total > int64(page*size)}, nil
	// 结束当前声明或代码作用域。
}

// 清空当前用户的收藏并返回数量。
func (s *Service) ClearFavorites(ctx context.Context, uid uint64) (int64, error) {
	// 清空当前用户的收藏并返回数量，把结果或错误返回给调用方。
	return s.Repo.ClearFavorites(ctx, uid)
	// 结束当前声明或代码作用域。
}
