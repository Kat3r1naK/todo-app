package repository

import (
	"context"

	"todo/server/internal/domain"
)

type UserRepository interface {
	FindByID(ctx context.Context, id uint) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	FindByAutologinKey(ctx context.Context, key string) (*domain.User, error)
	List(ctx context.Context) ([]domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	UpdatePassword(ctx context.Context, id uint, passwordHash string) error
	UpdateAutologinKey(ctx context.Context, id uint, key string) error
	Delete(ctx context.Context, id uint) error
	CountAdmins(ctx context.Context) (int64, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
}

type TaskRepository interface {
	ListByUser(ctx context.Context, userID uint) ([]domain.Task, error)
	FindByID(ctx context.Context, id, userID uint) (*domain.Task, error)
	Create(ctx context.Context, task *domain.Task) error
	Update(ctx context.Context, id, userID uint, updates map[string]any) error
	Delete(ctx context.Context, id, userID uint) error
	DeleteByUser(ctx context.Context, userID uint) error
	MaxPosition(ctx context.Context, userID uint, status string) (int, bool, error)
	MinPosition(ctx context.Context, userID uint, status string) (int, bool, error)
	Reorder(ctx context.Context, userID uint, columns map[string][]uint) error
	CountByUser(ctx context.Context) (map[uint]int64, error)
}

type NoteRepository interface {
	ListByUser(ctx context.Context, userID uint) ([]domain.Note, error)
	FindByID(ctx context.Context, id, userID uint) (*domain.Note, error)
	Create(ctx context.Context, note *domain.Note) error
	Update(ctx context.Context, id, userID uint, updates map[string]any) error
	Delete(ctx context.Context, id, userID uint) error
	DeleteByUser(ctx context.Context, userID uint) error
}
