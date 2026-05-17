package persistence

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"jcourse/config"
)

func NewPostgres(conf config.PostgresConfig) *gorm.DB {
	db, err := gorm.Open(postgres.Open(conf.DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	return db
}
