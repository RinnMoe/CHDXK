package persistence

import (
	"context"

	"github.com/redis/go-redis/v9"

	"jcourse/config"
)

func NewRedisClient(conf config.RedisConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Username: conf.Username,
		Password: conf.Password,
		DB:       conf.DB,
	})
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}
	return client
}
