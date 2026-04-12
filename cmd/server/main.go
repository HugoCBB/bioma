package main

import (
	"github.com/bioma/internal/config"
	"github.com/bioma/internal/handler"
)

func main() {
	_ = config.Load()
	handler.HandlerRequest()

}
