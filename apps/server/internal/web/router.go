package web

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"todo/server/internal/service"
)

func NewRouter(
	db *gorm.DB,
	auth *service.AuthService,
	authHandler *AuthHandler,
	taskHandler *TaskHandler,
	noteHandler *NoteHandler,
	adminHandler *AdminHandler,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Authorization"},
		MaxAge:          12 * time.Hour,
	}))

	r.GET("/healthz", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db-unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	m := newAuthMiddleware(auth)
	api := r.Group("/api")
	{
		api.POST("/auth/login", authHandler.Login)
		api.GET("/auth/autologin", authHandler.Autologin)

		authed := api.Group("", m.Required())
		{
			authed.GET("/auth/me", authHandler.Me)

			authed.GET("/tasks", taskHandler.List)
			authed.POST("/tasks", taskHandler.Create)
			authed.POST("/tasks/reorder", taskHandler.Reorder)
			authed.PATCH("/tasks/:id", taskHandler.Update)
			authed.DELETE("/tasks/:id", taskHandler.Delete)

			authed.GET("/notes", noteHandler.List)
			authed.POST("/notes", noteHandler.Create)
			authed.PATCH("/notes/:id", noteHandler.Update)
			authed.DELETE("/notes/:id", noteHandler.Delete)

			admin := authed.Group("/admin", m.AdminOnly())
			{
				admin.GET("/users", adminHandler.ListUsers)
				admin.POST("/users", adminHandler.CreateUser)
				admin.POST("/users/:id/password", adminHandler.ResetPassword)
				admin.DELETE("/users/:id", adminHandler.DeleteUser)
			}
		}
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "接口不存在"})
	})

	return r
}
