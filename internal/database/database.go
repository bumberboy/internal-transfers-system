package database

import (
	"fmt"
	"github.com/avast/retry-go/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"internal-transfers-system/config"
	"log"
	"log/slog"
	"os"
	"time"
)

func NewLogger() logger.Interface {
	return logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second * 20, // Slow SQL threshold
			LogLevel:                  logger.Error,     // Log level
			IgnoreRecordNotFoundError: true,             // Ignore ErrRecordNotFound error for logger
		},
	)
}

func NewDefaultDBClientOrFatal(config config.Config) *gorm.DB {

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Singapore",
		config.DBHost, config.DBUser, config.DBPassword, config.DBName, config.DBPort)

	slog.Debug("prepped dsn for db connection", "dsn", dsn)
	db, err := NewDBClient(dsn)
	if err != nil {
		slog.Error("failed to create new db client", err)
	}
	return db
}

func NewDBClient(dsn string) (*gorm.DB, error) {
	var odb *gorm.DB
	if err := retry.Do(
		func() error {
			slog.Info("connecting to db...")
			db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
				Logger: NewLogger(),
			})
			if err != nil {
				return err
			}
			sqlDB, err := db.DB()
			if err != nil {
				panic(err)
			}
			sqlDB.SetMaxOpenConns(100)
			sqlDB.SetMaxIdleConns(10)
			odb = db
			return nil
		},
		retry.Attempts(20),
	); err != nil {
		return nil, fmt.Errorf("connecting to db: %w", err)
	}

	return odb, nil
}
