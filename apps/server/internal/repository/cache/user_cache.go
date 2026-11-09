package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gomodule/redigo/redis"

	"todo/server/internal/domain"
)

const userCacheTTL = 10 * time.Minute

type UserCache struct {
	pool *redis.Pool
}

func NewUserCache(pool *redis.Pool) *UserCache {
	return &UserCache{pool: pool}
}

type cachedUser struct {
	ID           uint      `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"password_hash"`
	IsAdmin      bool      `json:"is_admin"`
	AutologinKey string    `json:"autologin_key"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func toCachedUser(user *domain.User) cachedUser {
	return cachedUser{
		ID:           user.ID,
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		IsAdmin:      user.IsAdmin,
		AutologinKey: user.AutologinKey,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

func (u cachedUser) toDomain() *domain.User {
	return &domain.User{
		ID:           u.ID,
		Username:     u.Username,
		PasswordHash: u.PasswordHash,
		IsAdmin:      u.IsAdmin,
		AutologinKey: u.AutologinKey,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

func (c *UserCache) GetByID(ctx context.Context, id uint) (*domain.User, bool) {
	return c.get(ctx, userIDKey(id))
}

func (c *UserCache) GetByUsername(ctx context.Context, username string) (*domain.User, bool) {
	return c.get(ctx, usernameKey(username))
}

func (c *UserCache) GetByAutologinKey(ctx context.Context, key string) (*domain.User, bool) {
	return c.get(ctx, autologinKey(key))
}

func (c *UserCache) SetByID(ctx context.Context, user *domain.User) {
	c.set(ctx, userIDKey(user.ID), user)
}

func (c *UserCache) SetByUsername(ctx context.Context, user *domain.User) {
	c.set(ctx, usernameKey(user.Username), user)
}

func (c *UserCache) SetByAutologinKey(ctx context.Context, user *domain.User) {
	if user.AutologinKey == "" {
		return
	}
	c.set(ctx, autologinKey(user.AutologinKey), user)
}

func (c *UserCache) DeleteByID(ctx context.Context, id uint) {
	c.del(ctx, userIDKey(id))
}

func (c *UserCache) DeleteByUsername(ctx context.Context, username string) {
	c.del(ctx, usernameKey(username))
}

func (c *UserCache) DeleteByAutologinKey(ctx context.Context, key string) {
	if key == "" {
		return
	}
	c.del(ctx, autologinKey(key))
}

func (c *UserCache) get(ctx context.Context, key string) (*domain.User, bool) {
	if c == nil || c.pool == nil {
		return nil, false
	}
	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		log.Printf("获取 Redis 连接失败: %v", err)
		return nil, false
	}
	defer conn.Close()

	data, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if !errors.Is(err, redis.ErrNil) {
			log.Printf("读取用户缓存失败: %v", err)
		}
		return nil, false
	}
	var value cachedUser
	if err := json.Unmarshal(data, &value); err != nil {
		log.Printf("解析用户缓存失败: %v", err)
		return nil, false
	}
	return value.toDomain(), true
}

func (c *UserCache) set(ctx context.Context, key string, user *domain.User) {
	if c == nil || c.pool == nil {
		return
	}
	data, err := json.Marshal(toCachedUser(user))
	if err != nil {
		log.Printf("序列化用户缓存失败: %v", err)
		return
	}
	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		log.Printf("获取 Redis 连接失败: %v", err)
		return
	}
	defer conn.Close()
	if _, err := conn.Do("SET", key, data, "EX", int(userCacheTTL.Seconds())); err != nil {
		log.Printf("写入用户缓存失败: %v", err)
	}
}

func (c *UserCache) del(ctx context.Context, keys ...string) {
	if c == nil || c.pool == nil {
		return
	}
	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		log.Printf("获取 Redis 连接失败: %v", err)
		return
	}
	defer conn.Close()
	args := make([]any, len(keys))
	for i, key := range keys {
		args[i] = key
	}
	if _, err := conn.Do("DEL", args...); err != nil {
		log.Printf("删除用户缓存失败: %v", err)
	}
}

func userIDKey(id uint) string {
	return fmt.Sprintf("todo:user:id:%d", id)
}

func usernameKey(username string) string {
	return "todo:user:name:" + username
}

func autologinKey(key string) string {
	return "todo:user:key:" + key
}
