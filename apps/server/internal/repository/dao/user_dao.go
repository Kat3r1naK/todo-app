package dao

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"todo/server/internal/domain"
	"todo/server/internal/errs"
)

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{db: db}
}

func (d *UserDAO) FindByID(ctx context.Context, id uint) (*domain.User, error) {
	var user domain.User
	if err := d.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, wrapUserError(err)
	}
	return &user, nil
}

func (d *UserDAO) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	if err := d.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, wrapUserError(err)
	}
	return &user, nil
}

func (d *UserDAO) FindByAutologinKey(ctx context.Context, key string) (*domain.User, error) {
	var user domain.User
	if err := d.db.WithContext(ctx).Where("autologin_key = ?", key).First(&user).Error; err != nil {
		return nil, wrapUserError(err)
	}
	return &user, nil
}

func (d *UserDAO) List(ctx context.Context) ([]domain.User, error) {
	var users []domain.User
	if err := d.db.WithContext(ctx).Order("id ASC").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (d *UserDAO) Create(ctx context.Context, user *domain.User) error {
	return d.db.WithContext(ctx).Create(user).Error
}

func (d *UserDAO) UpdatePassword(ctx context.Context, id uint, passwordHash string) error {
	return d.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", id).
		Update("password_hash", passwordHash).Error
}

func (d *UserDAO) UpdateAutologinKey(ctx context.Context, id uint, key string) error {
	return d.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", id).
		Update("autologin_key", key).Error
}

func (d *UserDAO) Delete(ctx context.Context, id uint) error {
	res := d.db.WithContext(ctx).Delete(&domain.User{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errs.NotFound("账号不存在")
	}
	return nil
}

func (d *UserDAO) CountAdmins(ctx context.Context) (int64, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&domain.User{}).Where("is_admin = ?", true).Count(&count).Error
	return count, err
}

func (d *UserDAO) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&domain.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

func wrapUserError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.NotFound("账号不存在")
	}
	return err
}
