package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type (
	redisClient struct {
		addr string
	}

	LoadEnv struct {
		RedisEnv redisClient
	}
)

func Load() *LoadEnv {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Erro ao carregar as variaveis de ambiente %v", err)
		return nil
	}

	redisClient := &redisClient{
		addr: os.Getenv("REDIS_ADDR"),
	}
	return &LoadEnv{
		RedisEnv: *redisClient,
	}
}
