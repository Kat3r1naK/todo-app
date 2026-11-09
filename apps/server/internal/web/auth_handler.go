package web

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"todo/server/internal/errs"
	"todo/server/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		abortWithError(c, errs.BadRequest("请求格式错误"))
		return
	}
	token, user, err := h.auth.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		abortWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": toUserDTO(user)})
}

func (h *AuthHandler) Autologin(c *gin.Context) {
	token, user, err := h.auth.Autologin(c.Request.Context(), c.Query("key"))
	if err != nil {
		abortWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": toUserDTO(user)})
}

func (h *AuthHandler) Me(c *gin.Context) {
	user := currentUser(c)
	c.JSON(http.StatusOK, gin.H{"user": toUserDTO(user)})
}
