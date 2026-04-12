package infra

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

func NewClientRedis(ctx context.Context) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Erro ao se conectar ao redis")
	}
	return rdb
}
