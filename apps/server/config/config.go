package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port          string
	DSN           string
	JWTSecret     string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:          strings.TrimSpace(os.Getenv("PORT")),
		DSN:           strings.TrimSpace(os.Getenv("DB_DSN")),
		JWTSecret:     strings.TrimSpace(os.Getenv("JWT_SECRET")),
		RedisAddr:     strings.TrimSpace(os.Getenv("REDIS_ADDR")),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       parseRedisDB(os.Getenv("REDIS_DB")),
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.DSN == "" {
		return nil, errors.New("环境变量 DB_DSN 未配置")
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "todo-dev-secret-change-me"
	}
	return cfg, nil
}

func parseRedisDB(raw string) int {
	db, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || db < 0 {
		return 0
	}
	return db
}
