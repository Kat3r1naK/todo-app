package web

import (
	"time"

	"todo/server/internal/domain"
)

type userDTO struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
}

func toUserDTO(user *domain.User) userDTO {
	return userDTO{
		ID:        user.ID,
		Username:  user.Username,
		IsAdmin:   user.IsAdmin,
		CreatedAt: user.CreatedAt,
	}
}

type adminUserDTO struct {
	ID           uint      `json:"id"`
	Username     string    `json:"username"`
	IsAdmin      bool      `json:"is_admin"`
	AutologinKey string    `json:"autologin_key"`
	TaskCount    int64     `json:"task_count"`
	CreatedAt    time.Time `json:"created_at"`
}
