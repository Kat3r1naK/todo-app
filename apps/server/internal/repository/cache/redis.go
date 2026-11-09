package cache

import (
	"log"
	"time"

	"github.com/gomodule/redigo/redis"

	"todo/server/config"
)

func NewRedisClient(cfg *config.Config) (*redis.Pool, func(), error) {
	if cfg.RedisAddr == "" {
		log.Println("未配置 REDIS_ADDR，用户缓存已禁用")
		return nil, func() {}, nil
	}
	pool := &redis.Pool{
		MaxIdle:     8,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", cfg.RedisAddr,
				redis.DialPassword(cfg.RedisPassword),
				redis.DialDatabase(cfg.RedisDB),
				redis.DialConnectTimeout(3*time.Second),
				redis.DialReadTimeout(3*time.Second),
				redis.DialWriteTimeout(3*time.Second),
			)
		},
		TestOnBorrow: func(conn redis.Conn, lastUsed time.Time) error {
			if time.Since(lastUsed) < time.Minute {
				return nil
			}
			_, err := conn.Do("PING")
			return err
		},
	}
	conn := pool.Get()
	defer conn.Close()
	if _, err := conn.Do("PING"); err != nil {
		log.Printf("Redis 暂不可用，缓存将降级为直连数据库: %v", err)
	}
	cleanup := func() {
		_ = pool.Close()
	}
	return pool, cleanup, nil
}
