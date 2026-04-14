package handler

import (
	"net/http"

	"github.com/bioma/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
)

func HandlerRequest(rdb *redis.Client, cfg *oauth2.Config) {
	r := gin.Default()
	repo := repository.NewRepository(rdb)
	authHandler := NewAuthHandlerSetup(repo, cfg)
	chatHandler := NewChatHandlerSetup(repo)

	api := r.Group("")
	{
		api.GET("", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{
				"health": "ok",
			})
		})

		auth := api.Group("/auth")
		{
			auth.GET("/login", authHandler.handleLogin)
			auth.GET("/callback", authHandler.handleCallback)
		}

		chat := api.Group("/chat")
		{
			chat.POST("/", chatHandler.handlerChat)
		}
	}
	r.Run(":8080")
}
