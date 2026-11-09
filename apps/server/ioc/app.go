package ioc

import (
	"github.com/gin-gonic/gin"

	"todo/server/internal/service"
)

type App struct {
	Router    *gin.Engine
	Bootstrap *service.Bootstrap
}

func NewApp(router *gin.Engine, bootstrap *service.Bootstrap) *App {
	return &App{Router: router, Bootstrap: bootstrap}
}
