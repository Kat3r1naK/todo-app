package web

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"todo/server/internal/domain"
	"todo/server/internal/service"
)

const userContextKey = "auth.user"

type authMiddleware struct {
	auth *service.AuthService
}

func newAuthMiddleware(auth *service.AuthService) *authMiddleware {
	return &authMiddleware{auth: auth}
}

func (m *authMiddleware) Required() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(c.GetHeader("Authorization"))
		if strings.HasPrefix(token, "Bearer ") {
			token = strings.TrimSpace(token[7:])
		}
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		user, err := m.auth.Authenticate(c.Request.Context(), token)
		if err != nil {
			abortWithError(c, err)
			return
		}
		c.Set(userContextKey, user)
		c.Next()
	}
}

func (m *authMiddleware) AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := currentUser(c)
		if user == nil || !user.IsAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "无权限"})
			return
		}
		c.Next()
	}
}

func currentUser(c *gin.Context) *domain.User {
	value, ok := c.Get(userContextKey)
	if !ok {
		return nil
	}
	user, _ := value.(*domain.User)
	return user
}
