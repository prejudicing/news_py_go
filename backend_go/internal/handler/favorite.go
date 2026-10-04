package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/prejudicing/news_py_go/backend_go/internal/dto"
	"github.com/prejudicing/news_py_go/backend_go/internal/service"
)

// 校验新闻 ID 并返回收藏状态。
func (s *Handler) checkFavorite(c *gin.Context) {
	id, valid := number(c, "newsId", -1, 1, 2147483647)
	if !valid {
		return
	}
	found, err := s.favorite.CheckFavorite(c.Request.Context(), currentUser(c).ID, uint64(id))
	s.respond(c, "检查收藏状态成功", gin.H{"isFavorite": found}, err)
}

// 绑定新闻 ID 并调用收藏服务。
func (s *Handler) addFavorite(c *gin.Context) {
	var input dto.NewsRequest
	if !bind(c, &input) {
		return
	}
	item, err := s.favorite.AddFavorite(c.Request.Context(), currentUser(c).ID, input.NewsID)
	s.respond(c, "添加收藏成功", gin.H{"id": item.ID, "user_id": item.UserID, "news_id": item.NewsID, "created_at": service.Timestamp(item.CreatedAt)}, err)
}

// 校验新闻 ID 并调用删除服务。
func (s *Handler) removeFavorite(c *gin.Context) {
	id, valid := number(c, "newsId", -1, 1, 2147483647)
	if !valid {
		return
	}
	err := s.favorite.RemoveFavorite(c.Request.Context(), currentUser(c).ID, uint64(id))
	s.respond(c, "删除收藏成功", nil, err)
}

// 校验页码并调用收藏列表服务。
func (s *Handler) favoriteList(c *gin.Context) {
	page, size, valid := pagination(c)
	if !valid {
		return
	}
	data, err := s.favorite.Favorites(c.Request.Context(), currentUser(c).ID, page, size)
	s.respond(c, "获取收藏列表成功", data, err)
}

// 调用收藏清空服务并返回删除数量。
func (s *Handler) clearFavorites(c *gin.Context) {
	count, err := s.favorite.ClearFavorites(c.Request.Context(), currentUser(c).ID)
	s.respond(c, fmt.Sprintf("清空了%d条记录", count), nil, err)
}
