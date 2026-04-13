package infra

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

func NewClientRedis(ctx context.Context) *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Erro ao se conectar ao redis: %v", err)
	}
	return rdb
}
