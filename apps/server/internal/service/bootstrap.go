package service

import (
	"context"
	"errors"
	"log"

	"todo/server/internal/domain"
	"todo/server/internal/errs"
	"todo/server/internal/repository"
	"todo/server/pkg/password"
)

type Bootstrap struct {
	users repository.UserRepository
}

func NewBootstrap(users repository.UserRepository) *Bootstrap {
	return &Bootstrap{users: users}
}

func (b *Bootstrap) Run(ctx context.Context) error {
	admin, err := b.users.FindByUsername(ctx, "admin")
	if err == nil {
		if admin.AutologinKey == "" {
			return b.users.UpdateAutologinKey(ctx, admin.ID, domain.NewAutologinKey())
		}
		return nil
	}
	if !errors.Is(err, errs.ErrNotFound) {
		return err
	}
	hash, err := password.Hash("admin")
	if err != nil {
		return err
	}
	admin = &domain.User{
		Username:     "admin",
		PasswordHash: hash,
		IsAdmin:      true,
		AutologinKey: domain.NewAutologinKey(),
	}
	if err := b.users.Create(ctx, admin); err != nil {
		return err
	}
	log.Println("已创建默认管理员账号: admin / admin")
	return nil
}
