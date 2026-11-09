package service

import (
	"context"
	"regexp"
	"strings"

	"todo/server/internal/domain"
	"todo/server/internal/errs"
	"todo/server/internal/repository"
	"todo/server/pkg/password"
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,32}$`)

type CreateUserInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"is_admin"`
}

type UserWithStats struct {
	domain.User
	TaskCount int64
}

type UserService struct {
	users repository.UserRepository
	tasks repository.TaskRepository
	notes repository.NoteRepository
}

func NewUserService(users repository.UserRepository, tasks repository.TaskRepository, notes repository.NoteRepository) *UserService {
	return &UserService{users: users, tasks: tasks, notes: notes}
}

func (s *UserService) List(ctx context.Context) ([]UserWithStats, error) {
	users, err := s.users.List(ctx)
	if err != nil {
		return nil, err
	}
	counts, err := s.tasks.CountByUser(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]UserWithStats, 0, len(users))
	for _, user := range users {
		result = append(result, UserWithStats{User: user, TaskCount: counts[user.ID]})
	}
	return result, nil
}

func (s *UserService) Create(ctx context.Context, input CreateUserInput) error {
	username := strings.TrimSpace(input.Username)
	if !usernamePattern.MatchString(username) {
		return errs.BadRequest("账号需为 1-32 位字母、数字、下划线或短横线")
	}
	if input.Password == "" || len(input.Password) > 72 {
		return errs.BadRequest("密码长度需为 1-72 位")
	}
	exists, err := s.users.ExistsByUsername(ctx, username)
	if err != nil {
		return err
	}
	if exists {
		return errs.Conflict("账号已存在")
	}
	hash, err := password.Hash(input.Password)
	if err != nil {
		return err
	}
	user := &domain.User{
		Username:     username,
		PasswordHash: hash,
		IsAdmin:      input.IsAdmin,
		AutologinKey: domain.NewAutologinKey(),
	}
	return s.users.Create(ctx, user)
}

func (s *UserService) ResetPassword(ctx context.Context, userID uint, plainPassword string) error {
	if _, err := s.users.FindByID(ctx, userID); err != nil {
		return err
	}
	if plainPassword == "" || len(plainPassword) > 72 {
		return errs.BadRequest("密码长度需为 1-72 位")
	}
	hash, err := password.Hash(plainPassword)
	if err != nil {
		return err
	}
	return s.users.UpdatePassword(ctx, userID, hash)
}

func (s *UserService) Delete(ctx context.Context, currentUserID, targetUserID uint) error {
	target, err := s.users.FindByID(ctx, targetUserID)
	if err != nil {
		return err
	}
	if target.ID == currentUserID {
		return errs.BadRequest("不能删除当前登录账号")
	}
	if target.IsAdmin {
		count, err := s.users.CountAdmins(ctx)
		if err != nil {
			return err
		}
		if count <= 1 {
			return errs.BadRequest("至少需要保留一个管理员")
		}
	}
	if err := s.tasks.DeleteByUser(ctx, targetUserID); err != nil {
		return err
	}
	if err := s.notes.DeleteByUser(ctx, targetUserID); err != nil {
		return err
	}
	return s.users.Delete(ctx, targetUserID)
}
