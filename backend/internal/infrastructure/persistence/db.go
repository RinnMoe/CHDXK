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
	DSN          string `mapstructure:"dsn"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

func NewPostgres(conf PostgresConfig) *gorm.DB {
	db, err := gorm.Open(postgres.Open(conf.DSN), &gorm.Config{
		Logger: NewGormLogger(),
	})
	if err != nil {
		panic(err)
	}
	pool, err := db.DB()
	if err != nil {
		panic(err)
	}
	maxOpen := conf.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = 6
	}
	maxIdle := conf.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = 2
	}
	if maxIdle > maxOpen {
		maxIdle = maxOpen
	}
	pool.SetMaxOpenConns(maxOpen)
	pool.SetMaxIdleConns(maxIdle)
	pool.SetConnMaxIdleTime(5 * time.Minute)

	return db
}
