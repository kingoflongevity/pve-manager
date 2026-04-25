package database

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/kingoflongevity/pve-manager/backend/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDatabase(log *zap.Logger) (*gorm.DB, error) {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/pve_manager.db"
	}

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据库目录失败: %w", err)
	}

	dsn := dbPath
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.New(&zapLogger{log: log}, logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}

	DB = db
	log.Info("SQLite 数据库初始化成功", zap.String("path", dbPath))
	return db, nil
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.PVEConfig{},
		&model.UserSession{},
		&model.AuditLog{},
		&model.SystemConfig{},
	)
}

type zapLogger struct {
	log *zap.Logger
}

func (l *zapLogger) Printf(format string, args ...interface{}) {
	l.log.Sugar().Infof(format, args...)
}

func GetDB() *gorm.DB {
	if DB == nil {
		panic("数据库未初始化，请先调用 InitDatabase")
	}
	return DB
}
