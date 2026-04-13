package main

import (
	"context"

	"github.com/bioma/internal/config"
	"github.com/bioma/internal/handler"
	"github.com/bioma/internal/infra"
)

func init() {
	config.LoadEnv()
}

func main() {
	ctx := context.Background()
	rdb := infra.NewClientRedis(ctx)
	handler.HandlerRequest(rdb)
}
