package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/prejudicing/news_py_go/backend_go/internal/dto"
	"github.com/prejudicing/news_py_go/backend_go/internal/service"
)

// 绑定新闻 ID 并调用历史写入服务。
func (s *Handler) addHistory(c *gin.Context) {
	var input dto.NewsRequest
	if !bind(c, &input) {
		return
	}
	item, err := s.history.AddHistory(c.Request.Context(), currentUser(c).ID, input.NewsID)
	s.respond(c, "添加成功", gin.H{"id": item.ID, "user_id": item.UserID, "news_id": item.NewsID, "view_time": service.Timestamp(item.ViewTime)}, err)
}

// 校验页码并调用历史列表服务。
func (s *Handler) historyList(c *gin.Context) {
	page, size, valid := pagination(c)
	if !valid {
		return
	}
	data, err := s.history.Histories(c.Request.Context(), currentUser(c).ID, page, size)
	s.respond(c, "success", data, err)
}

// 校验记录主键并调用历史删除服务。
func (s *Handler) deleteHistory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("history_id"), 10, 64)
	if err != nil || id == 0 {
		fail(c, 422, "history_id必须为正整数")
		return
	}
	err = s.history.DeleteHistory(c.Request.Context(), currentUser(c).ID, id)
	s.respond(c, "删除成功", nil, err)
}

// 调用历史清空服务。
func (s *Handler) clearHistory(c *gin.Context) {
	err := s.history.ClearHistory(c.Request.Context(), currentUser(c).ID)
	s.respond(c, "清空成功", nil, err)
}
