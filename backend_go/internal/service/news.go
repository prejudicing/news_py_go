package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/prejudicing/news_py_go/backend_go/internal/dto"
	"github.com/prejudicing/news_py_go/backend_go/internal/model"
	"github.com/prejudicing/news_py_go/backend_go/internal/repository"
)

// 读取排序后的新闻分类。
func (s *NewsService) Categories(ctx context.Context, skip, limit int) ([]model.Category, error) {
	key := fmt.Sprintf("categories:%d:%d", skip, limit)
	items := make([]model.Category, 0)
	if !s.Cache.Get(ctx, key, &items) {
		var err error
		items, err = s.Repo.Categories(ctx, skip, limit)
		if err != nil {
			return nil, err
		}
		s.Cache.Set(ctx, key, items, 2*time.Hour)
	}
	return items, nil
}

// 读取指定分类的分页新闻与总数。
func (s *NewsService) NewsList(ctx context.Context, category, page, size int) (dto.Page[dto.NewsItem], error) {
	total, err := s.Repo.CountNews(ctx, category)
	if err != nil {
		return dto.Page[dto.NewsItem]{}, err
	}
	items := make([]dto.NewsItem, 0)
	key := fmt.Sprintf("list:%d:%d:%d", category, page, size)
	if !s.Cache.Get(ctx, key, &items) {
		rows, err := s.Repo.ListNews(ctx, category, (page-1)*size, size)
		if err != nil {
			return dto.Page[dto.NewsItem]{}, err
		}
		for _, row := range rows {
			items = append(items, newsDTO(row))
		}
		s.Cache.Set(ctx, key, items, 30*time.Minute)
	}
	return dto.Page[dto.NewsItem]{List: items, Total: total, HasMore: int64((page-1)*size+len(items)) < total}, nil
}

// 读取新闻详情、累计浏览量并查询相关新闻。
func (s *NewsService) NewsDetail(ctx context.Context, id uint64) (dto.NewsDetail, error) {
	var item model.News
	key := fmt.Sprintf("detail:%d", id)
	if !s.Cache.Get(ctx, key, &item) {
		var err error
		item, err = s.Repo.NewsByID(ctx, id)
		if errors.Is(err, repository.ErrNotFound) {
			return dto.NewsDetail{}, problem(ErrorNotFound, "新闻不存在")
		}
		if err != nil {
			return dto.NewsDetail{}, err
		}
		s.Cache.Set(ctx, key, item, 5*time.Minute)
	}
	views, err := s.Repo.IncreaseViews(ctx, id)
	if err != nil {
		return dto.NewsDetail{}, err
	}
	item.Views = views
	related := make([]dto.RelatedItem, 0)
	relatedKey := fmt.Sprintf("related:%d:%d", item.ID, item.CategoryID)
	if !s.Cache.Get(ctx, relatedKey, &related) {
		rows, err := s.Repo.RelatedNews(ctx, item.ID, item.CategoryID)
		if err != nil {
			return dto.NewsDetail{}, err
		}
		for _, row := range rows {
			related = append(related, dto.RelatedItem{ID: row.ID, Title: row.Title, Image: row.Image, Views: row.Views})
		}
		s.Cache.Set(ctx, relatedKey, related, 30*time.Minute)
	}
	return dto.NewsDetail{ID: item.ID, Title: item.Title, Content: item.Content, Image: item.Image, Author: item.Author, PublishTime: Timestamp(item.PublishTime), CategoryID: item.CategoryID, Views: item.Views, RelatedNews: related}, nil
}
