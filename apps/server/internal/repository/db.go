package repository

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"todo/server/config"
	"todo/server/internal/domain"
)

func openAndPing(dsn string) (*gorm.DB, error) {
	dbLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: dbLogger})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	return db, nil
}

func NewDB(cfg *config.Config) (*gorm.DB, func(), error) {
	var db *gorm.DB
	var err error
	for i := 1; i <= 30; i++ {
		db, err = openAndPing(cfg.DSN)
		if err == nil {
			break
		}
		if i < 30 {
			log.Printf("数据库未就绪 (%d/30)，2 秒后重试: %v", i, err)
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("数据库连接失败: %w", err)
	}
	if err := db.AutoMigrate(&domain.User{}, &domain.Task{}, &domain.Note{}); err != nil {
		return nil, nil, fmt.Errorf("数据表迁移失败: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() {
		_ = sqlDB.Close()
	}
	return db, cleanup, nil
}
