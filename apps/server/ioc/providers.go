package ioc

import (
	"github.com/google/wire"

	"todo/server/config"
	"todo/server/internal/repository"
	"todo/server/internal/repository/cache"
	"todo/server/internal/repository/dao"
	"todo/server/internal/service"
	"todo/server/internal/web"
	"todo/server/pkg/jwt"
)

func provideTokenManager(cfg *config.Config) *jwt.Manager {
	return jwt.NewManager(cfg.JWTSecret)
}

var ProviderSet = wire.NewSet(
	repository.NewDB,
	cache.NewRedisClient,
	cache.NewUserCache,
	dao.NewUserDAO,
	dao.NewTaskDAO,
	dao.NewNoteDAO,
	repository.NewUserRepository,
	repository.NewTaskRepository,
	repository.NewNoteRepository,
	provideTokenManager,
	service.NewAuthService,
	service.NewTaskService,
	service.NewNoteService,
	service.NewUserService,
	service.NewBootstrap,
	web.NewAuthHandler,
	web.NewTaskHandler,
	web.NewNoteHandler,
	web.NewAdminHandler,
	web.NewRouter,
	NewApp,
)
