package main

import (
	"context"

	"github.com/bioma/internal/config"
	"github.com/bioma/internal/handler"
	"github.com/bioma/internal/infra"
)

func main() {
	ctx := context.Background()
	config.LoadEnv()

	googleCfg := config.NewGoogleClientSetup()
	rdb := infra.NewClientRedis(ctx)
	handler.HandlerRequest(rdb, googleCfg)
}
