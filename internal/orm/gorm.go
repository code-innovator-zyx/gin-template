package orm

import (
	"errors"
	"fmt"
	sqldns "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

// Config 数据库配置
type Config struct {
	DSN          string `mapstructure:"dsn" validate:"required"`
	MaxOpenConns int    `mapstructure:"max_open_conns" validate:"gte=1,lte=10000"`
	MaxIdleConns int    `mapstructure:"max_idle_conns" validate:"gte=0,lte=10000,ltefield=MaxOpenConns"`
	MaxLifetime  int    `mapstructure:"max_life_time" validate:"gte=60"` // 秒
	LogLevel     int    `mapstructure:"log_level" validate:"min=0,max=4"`
}

// Init 初始化数据库连接
func Init(c Config) (*gorm.DB, error) {
	// 先创建数据库（如果不存在）
	if err := createDatabaseIfNotExist(c); err != nil {
		return nil, err
	}

	gormConfig := &gorm.Config{
		Logger: getLogger(c.LogLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC() // 建议统一 UTC
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	db, err := gorm.Open(mysql.Open(c.DSN), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("gorm open 失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取原生 DB 失败: %w", err)
	}

	sqlDB.SetMaxOpenConns(c.MaxOpenConns)
	sqlDB.SetMaxIdleConns(c.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(c.MaxLifetime) * time.Second)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("数据库 ping 失败: %w", err)
	}

	logrus.Infof("数据库连接成功: %s", c.DSN)
	return db, nil
}

// createDatabaseIfNotExist 创建数据库，如果不存在
func createDatabaseIfNotExist(c Config) error {
	cfg, err := sqldns.ParseDSN(c.DSN)
	if err != nil {
		return fmt.Errorf("解析 DSN 失败: %w", err)
	}

	dbName := cfg.DBName
	baseDSN := fmt.Sprintf("%s:%s@tcp(%s)/", cfg.User, cfg.Passwd, cfg.Addr)

	sysDB, err := gorm.Open(mysql.Open(baseDSN), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("连接 MySQL 失败: %w", err)
	}

	createSQL := fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;",
		dbName,
	)
	return sysDB.Exec(createSQL).Error
}

// getLogger 根据日志级别生成 GORM Logger
func getLogger(logLevel int) logger.Interface {
	var lvl logger.LogLevel
	switch logLevel {
	case 0:
		lvl = logger.Silent
	case 1:
		lvl = logger.Error
	case 2:
		lvl = logger.Warn
	case 3:
		lvl = logger.Info
	case 4:
		lvl = logger.Info
	default:
		lvl = logger.Info
	}

	return logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  lvl,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)
}

// UpdateLogLevel 动态修改 GORM 日志等级
func UpdateLogLevel(db *gorm.DB, level int) error {
	if db == nil {
		return errors.New("db 为 nil")
	}
	newLogger := getLogger(level)
	db.Config.Logger = newLogger
	return nil
}
