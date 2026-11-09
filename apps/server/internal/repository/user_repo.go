package repository

import (
	"context"

	"todo/server/internal/domain"
	"todo/server/internal/repository/cache"
	"todo/server/internal/repository/dao"
)

type userRepository struct {
	dao   *dao.UserDAO
	cache *cache.UserCache
}

func NewUserRepository(userDAO *dao.UserDAO, userCache *cache.UserCache) UserRepository {
	return &userRepository{dao: userDAO, cache: userCache}
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*domain.User, error) {
	if user, ok := r.cache.GetByID(ctx, id); ok {
		return user, nil
	}
	user, err := r.dao.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	r.warm(ctx, user)
	return user, nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	if user, ok := r.cache.GetByUsername(ctx, username); ok {
		return user, nil
	}
	user, err := r.dao.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	r.warm(ctx, user)
	return user, nil
}

func (r *userRepository) FindByAutologinKey(ctx context.Context, key string) (*domain.User, error) {
	if user, ok := r.cache.GetByAutologinKey(ctx, key); ok {
		return user, nil
	}
	user, err := r.dao.FindByAutologinKey(ctx, key)
	if err != nil {
		return nil, err
	}
	r.warm(ctx, user)
	return user, nil
}

func (r *userRepository) List(ctx context.Context) ([]domain.User, error) {
	return r.dao.List(ctx)
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	if err := r.dao.Create(ctx, user); err != nil {
		return err
	}
	r.warm(ctx, user)
	return nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, id uint, passwordHash string) error {
	user, err := r.dao.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := r.dao.UpdatePassword(ctx, id, passwordHash); err != nil {
		return err
	}
	r.invalidate(ctx, user)
	return nil
}

func (r *userRepository) UpdateAutologinKey(ctx context.Context, id uint, key string) error {
	user, err := r.dao.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := r.dao.UpdateAutologinKey(ctx, id, key); err != nil {
		return err
	}
	r.invalidate(ctx, user)
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uint) error {
	user, err := r.dao.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := r.dao.Delete(ctx, id); err != nil {
		return err
	}
	r.invalidate(ctx, user)
	return nil
}

func (r *userRepository) CountAdmins(ctx context.Context) (int64, error) {
	return r.dao.CountAdmins(ctx)
}

func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return r.dao.ExistsByUsername(ctx, username)
}

func (r *userRepository) warm(ctx context.Context, user *domain.User) {
	r.cache.SetByID(ctx, user)
	r.cache.SetByUsername(ctx, user)
	r.cache.SetByAutologinKey(ctx, user)
}

func (r *userRepository) invalidate(ctx context.Context, user *domain.User) {
	r.cache.DeleteByID(ctx, user.ID)
	r.cache.DeleteByUsername(ctx, user.Username)
	r.cache.DeleteByAutologinKey(ctx, user.AutologinKey)
}
