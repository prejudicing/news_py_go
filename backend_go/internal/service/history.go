// 声明独立业务逻辑包。
package service

// 导入本文件使用的标准库、第三方依赖和内部包。
import (
	// 引入context，用于请求取消与超时上下文。
	"context"
	// 引入errors，用于错误类型匹配和构造。
	"errors"

	// 引入github.com/prejudicing/news_py_go/backend_go/internal/dto，用于接口参数和响应结构。
	"github.com/prejudicing/news_py_go/backend_go/internal/dto"
	// 引入github.com/prejudicing/news_py_go/backend_go/internal/model，用于数据库实体。
	"github.com/prejudicing/news_py_go/backend_go/internal/model"
	// 引入github.com/prejudicing/news_py_go/backend_go/internal/repository，用于仓储接口和统一错误。
	"github.com/prejudicing/news_py_go/backend_go/internal/repository"
	// 结束当前声明或代码作用域。
)

// 在用户行锁保护下创建或更新浏览记录。
func (s *Service) AddHistory(ctx context.Context, uid, nid uint64) (model.History, error) {
	// 声明当前实体。
	var item model.History
	// 执行事务回调，失败时回滚全部写入并保存操作结果。
	err := s.Repo.Transaction(ctx, func(tx repository.Store) error {
		// 检查本次操作是否失败，必要时提前返回错误。
		if _, err := tx.UserByID(ctx, uid, true); err != nil {
			// 返回当前结果与错误，由调用方决定响应或事务回滚。
			return err
			// 结束当前声明或代码作用域。
		}
		// 检查本次操作是否失败，必要时提前返回错误。
		if _, err := tx.NewsByID(ctx, nid); err != nil {
			// 返回当前结果与错误，由调用方决定响应或事务回滚。
			return err
			// 结束当前声明或代码作用域。
		}
		// 声明操作错误。
		var err error
		// 查询当前用户对指定新闻的最新历史记录并保存操作结果。
		item, err = tx.HistoryByNews(ctx, uid, nid)
		// 检查错误是否表示记录不存在，选择相应的业务分支。
		if errors.Is(err, repository.ErrNotFound) {
			// 获取应用时区中的当前时间并保存操作结果。
			item = model.History{UserID: uid, NewsID: nid, ViewTime: s.now()}
			// 写入新的浏览历史记录，把结果或错误返回给调用方。
			return tx.CreateHistory(ctx, &item)
			// 结束当前声明或代码作用域。
		}
		// 检查本次操作是否失败，必要时提前返回错误。
		if err != nil {
			// 返回当前结果与错误，由调用方决定响应或事务回滚。
			return err
			// 结束当前声明或代码作用域。
		}
		// 获取应用时区中的当前时间并保存操作结果。
		item.ViewTime = s.now()
		// 按历史记录主键更新浏览时间，把结果或错误返回给调用方。
		return tx.UpdateHistory(ctx, item.ID, item.ViewTime)
		// 结束当前回调或复合值构造。
	})
	// 返回当前结果与错误，由调用方决定响应或事务回滚。
	return item, err
	// 结束当前声明或代码作用域。
}

// 查询当前用户的历史分页列表。
func (s *Service) Histories(ctx context.Context, uid uint64, page, size int) (dto.Page[dto.HistoryItem], error) {
	// 查询当前用户的历史分页列表并保存操作结果。
	rows, total, err := s.Repo.Histories(ctx, uid, (page-1)*size, size)
	// 检查本次操作是否失败，必要时提前返回错误。
	if err != nil {
		// 返回当前结果与错误，由调用方决定响应或事务回滚。
		return dto.Page[dto.HistoryItem]{}, err
		// 结束当前声明或代码作用域。
	}
	// 创建非空列表容器，使无数据时 JSON 返回空数组。
	items := make([]dto.HistoryItem, 0, len(rows))
	// 遍历仓储查询结果，转换为前端响应结构。
	for _, row := range rows {
		// 把当前实体的响应字段加入列表。
		items = append(items, dto.HistoryItem{NewsItem: newsDTO(row.News), ViewTime: Timestamp(row.ViewTime), HistoryID: row.HistoryID})
		// 结束当前声明或代码作用域。
	}
	// 构造接口响应，返回数据和成功标记。
	return dto.Page[dto.HistoryItem]{List: items, Total: total, HasMore: total > int64(page*size)}, nil
	// 结束当前声明或代码作用域。
}

// 按记录主键删除当前用户的历史。
func (s *Service) DeleteHistory(ctx context.Context, uid, id uint64) error {
	// 按记录主键删除当前用户的历史并保存操作结果。
	count, err := s.Repo.DeleteHistory(ctx, uid, id)
	// 检查本次操作是否失败，必要时提前返回错误。
	if err != nil {
		// 返回当前结果与错误，由调用方决定响应或事务回滚。
		return err
		// 结束当前声明或代码作用域。
	}
	// 没有影响任何记录时返回记录不存在错误。
	if count == 0 {
		// 返回校验失败对应的业务状态码和消息。
		return problem(404, "历史记录不存在")
		// 结束当前声明或代码作用域。
	}
	// 返回空错误，表示操作成功。
	return nil
	// 结束当前声明或代码作用域。
}

// 清空当前用户的浏览历史。
func (s *Service) ClearHistory(ctx context.Context, uid uint64) error {
	// 清空当前用户的浏览历史并保存操作结果。
	_, err := s.Repo.ClearHistory(ctx, uid)
	// 返回当前结果与错误，由调用方决定响应或事务回滚。
	return err
	// 结束当前声明或代码作用域。
}
