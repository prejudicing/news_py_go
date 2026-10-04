package service

import (
	"context"
	"errors"

	"github.com/prejudicing/news_py_go/backend_go/internal/dto"
	"github.com/prejudicing/news_py_go/backend_go/internal/model"
	"github.com/prejudicing/news_py_go/backend_go/internal/repository"
)

// 在用户行锁保护下创建或更新浏览记录。
func (s *HistoryService) AddHistory(ctx context.Context, uid, nid uint64) (model.History, error) {
	var item model.History
	err := s.Repo.WithHistoryTransaction(ctx, func(tx repository.HistoryRepository) error {
		if _, err := tx.UserByID(ctx, uid, true); err != nil {
			return err
		}
		if _, err := tx.NewsByID(ctx, nid); err != nil {
			return err
		}
		var err error
		item, err = tx.HistoryByNews(ctx, uid, nid)
		if errors.Is(err, repository.ErrNotFound) {
			item = model.History{UserID: uid, NewsID: nid, ViewTime: s.now()}
			return tx.CreateHistory(ctx, &item)
		}
		if err != nil {
			return err
		}
		item.ViewTime = s.now()
		return tx.UpdateHistory(ctx, item.ID, item.ViewTime)
	})
	return item, err
}

// 查询当前用户的历史分页列表。
func (s *HistoryService) Histories(ctx context.Context, uid uint64, page, size int) (dto.Page[dto.HistoryItem], error) {
	rows, total, err := s.Repo.Histories(ctx, uid, (page-1)*size, size)
	if err != nil {
		return dto.Page[dto.HistoryItem]{}, err
	}
	items := make([]dto.HistoryItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.HistoryItem{NewsItem: newsDTO(row.News), ViewTime: Timestamp(row.ViewTime), HistoryID: row.HistoryID})
	}
	return dto.Page[dto.HistoryItem]{List: items, Total: total, HasMore: total > int64(page*size)}, nil
}

// 按记录主键删除当前用户的历史。
func (s *HistoryService) DeleteHistory(ctx context.Context, uid, id uint64) error {
	count, err := s.Repo.DeleteHistory(ctx, uid, id)
	if err != nil {
		return err
	}
	if count == 0 {
		return problem(404, "历史记录不存在")
	}
	return nil
}

// 清空当前用户的浏览历史。
func (s *HistoryService) ClearHistory(ctx context.Context, uid uint64) error {
	_, err := s.Repo.ClearHistory(ctx, uid)
	return err
}
