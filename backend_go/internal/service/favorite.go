package service

import (
	"context"

	"github.com/prejudicing/news_py_go/backend_go/internal/dto"
	"github.com/prejudicing/news_py_go/backend_go/internal/model"
)

// 检查当前用户是否收藏指定新闻。
func (s *FavoriteService) CheckFavorite(ctx context.Context, uid, nid uint64) (bool, error) {
	return s.Repo.FavoriteExists(ctx, uid, nid)
}

// 检查新闻存在后添加当前用户的收藏。
func (s *FavoriteService) AddFavorite(ctx context.Context, uid, nid uint64) (model.Favorite, error) {
	if _, err := s.Repo.NewsByID(ctx, nid); err != nil {
		return model.Favorite{}, err
	}
	item := model.Favorite{UserID: uid, NewsID: nid, CreatedAt: s.now()}
	err := s.Repo.CreateFavorite(ctx, &item)
	return item, err
}

// 删除当前用户的指定收藏并检查删除结果。
func (s *FavoriteService) RemoveFavorite(ctx context.Context, uid, nid uint64) error {
	count, err := s.Repo.DeleteFavorite(ctx, uid, nid)
	if err != nil {
		return err
	}
	if count == 0 {
		return problem(404, "收藏记录不存在")
	}
	return nil
}

// 查询当前用户的收藏分页列表。
func (s *FavoriteService) Favorites(ctx context.Context, uid uint64, page, size int) (dto.Page[dto.FavoriteItem], error) {
	rows, total, err := s.Repo.Favorites(ctx, uid, (page-1)*size, size)
	if err != nil {
		return dto.Page[dto.FavoriteItem]{}, err
	}
	items := make([]dto.FavoriteItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.FavoriteItem{NewsItem: newsDTO(row.News), FavoriteTime: Timestamp(row.FavoriteTime), FavoriteID: row.FavoriteID})
	}
	return dto.Page[dto.FavoriteItem]{List: items, Total: total, HasMore: total > int64(page*size)}, nil
}

// 清空当前用户的收藏并返回数量。
func (s *FavoriteService) ClearFavorites(ctx context.Context, uid uint64) (int64, error) {
	return s.Repo.ClearFavorites(ctx, uid)
}
