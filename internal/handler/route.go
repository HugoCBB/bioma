package handler

import (
	"net/http"

	"github.com/bioma/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func HandlerRequest(rdb *redis.Client) {
	r := gin.Default()
	repositoryAuth := repository.NewAuthRepository(rdb)
	authHandler := NewAuthHandlerSetup(repositoryAuth)
	api := r.Group("/")
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
	}
	r.Run(":8080")
}
