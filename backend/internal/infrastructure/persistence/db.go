package persistence

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresConfig struct {
	DSN string `mapstructure:"dsn"`
}

func NewPostgres(conf PostgresConfig) *gorm.DB {
	db, err := gorm.Open(postgres.Open(conf.DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	return db
}
