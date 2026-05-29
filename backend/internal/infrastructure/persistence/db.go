package persistence

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"jcourse/pkg/logx"
)

func NewGormLogger() logger.Interface {
	return logger.NewSlogLogger(logx.Logger(), logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  logger.Warn,
		IgnoreRecordNotFoundError: true,
		ParameterizedQueries:      false,
	})
}

type PostgresConfig struct {
	DSN string `mapstructure:"dsn"`
}

func NewPostgres(conf PostgresConfig) *gorm.DB {
	db, err := gorm.Open(postgres.Open(conf.DSN), &gorm.Config{
		Logger: NewGormLogger(),
	})
	if err != nil {
		panic(err)
	}

	return db
}
