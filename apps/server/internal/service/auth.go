package service

import (
	"context"
	"errors"
	"strings"

	"todo/server/internal/domain"
	"todo/server/internal/errs"
	"todo/server/internal/repository"
	"todo/server/pkg/jwt"
	"todo/server/pkg/password"
)

type AuthService struct {
	users  repository.UserRepository
	tokens *jwt.Manager
}

func NewAuthService(users repository.UserRepository, tokens *jwt.Manager) *AuthService {
	return &AuthService{users: users, tokens: tokens}
}

func (s *AuthService) Login(ctx context.Context, username, plainPassword string) (string, *domain.User, error) {
	username = strings.TrimSpace(username)
	if username == "" || plainPassword == "" {
		return "", nil, errs.BadRequest("请输入账号和密码")
	}
	user, err := s.users.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return "", nil, errs.InvalidCredentials("账号或密码错误")
		}
		return "", nil, err
	}
	if !password.Verify(user.PasswordHash, plainPassword) {
		return "", nil, errs.InvalidCredentials("账号或密码错误")
	}
	token, err := s.tokens.Generate(user.ID)
	if err != nil {
		return "", nil, err
	}
	return token, user, nil
}

func (s *AuthService) Autologin(ctx context.Context, key string) (string, *domain.User, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", nil, errs.BadRequest("缺少免登密钥")
	}
	user, err := s.users.FindByAutologinKey(ctx, key)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return "", nil, errs.Unauthorized("免登密钥无效")
		}
		return "", nil, err
	}
	token, err := s.tokens.Generate(user.ID)
	if err != nil {
		return "", nil, err
	}
	return token, user, nil
}

func (s *AuthService) Authenticate(ctx context.Context, token string) (*domain.User, error) {
	userID, err := s.tokens.Parse(token)
	if err != nil {
		return nil, errs.Unauthorized("登录已过期，请重新登录")
	}
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil, errs.Unauthorized("账号不存在")
		}
		return nil, err
	}
	return user, nil
}
