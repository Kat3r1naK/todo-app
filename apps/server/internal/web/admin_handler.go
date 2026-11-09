package web

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"todo/server/internal/errs"
	"todo/server/internal/service"
)

type AdminHandler struct {
	users *service.UserService
}

func NewAdminHandler(users *service.UserService) *AdminHandler {
	return &AdminHandler{users: users}
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	rows, err := h.users.List(c.Request.Context())
	if err != nil {
		abortWithError(c, err)
		return
	}
	users := make([]adminUserDTO, 0, len(rows))
	for _, row := range rows {
		users = append(users, adminUserDTO{
			ID:           row.ID,
			Username:     row.Username,
			IsAdmin:      row.IsAdmin,
			AutologinKey: row.AutologinKey,
			TaskCount:    row.TaskCount,
			CreatedAt:    row.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (h *AdminHandler) CreateUser(c *gin.Context) {
	var req service.CreateUserInput
	if err := c.ShouldBindJSON(&req); err != nil {
		abortWithError(c, errs.BadRequest("请求格式错误"))
		return
	}
	if err := h.users.Create(c.Request.Context(), req); err != nil {
		abortWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

func (h *AdminHandler) ResetPassword(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		abortWithError(c, errs.BadRequest("请求格式错误"))
		return
	}
	if err := h.users.ResetPassword(c.Request.Context(), id, req.Password); err != nil {
		abortWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.users.Delete(c.Request.Context(), currentUser(c).ID, id); err != nil {
		abortWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
