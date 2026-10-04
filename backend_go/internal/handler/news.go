package handler

import "github.com/gin-gonic/gin"

// 校验分类分页参数并调用业务服务。
func (s *Handler) categories(c *gin.Context) {
	skip, valid := number(c, "skip", 0, 0, 1000000)
	if !valid {
		return
	}
	limit, valid := number(c, "limit", 100, 1, 1000)
	if !valid {
		return
	}
	data, err := s.news.Categories(c.Request.Context(), skip, limit)
	s.respond(c, "获取新闻分类成功", data, err)
}

// 校验新闻分类与页码并调用业务服务。
func (s *Handler) newsList(c *gin.Context) {
	category, valid := number(c, "categoryId", -1, 1, 2147483647)
	if !valid {
		return
	}
	page, size, valid := pagination(c)
	if !valid {
		return
	}
	data, err := s.news.NewsList(c.Request.Context(), category, page, size)
	s.respond(c, "获取新闻列表成功", data, err)
}

// 校验新闻 ID 并调用详情服务。
func (s *Handler) newsDetail(c *gin.Context) {
	id, valid := number(c, "id", -1, 1, 2147483647)
	if !valid {
		return
	}
	data, err := s.news.NewsDetail(c.Request.Context(), uint64(id))
	s.respond(c, "success", data, err)
}
