package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandlerRequest() {
	r := gin.New()
	api := r.Group("/")
	{
		api.GET("", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{
				"healh": "ok",
			})
		})
	}
	r.Run(":8080")
}
