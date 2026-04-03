package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleRequest() {
	r := gin.Default()

	api := r.Group("/api")
	{
		api.GET("/", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{
				"status": "ok",
			})
		})
	}

	r.Run(":8080")
}
