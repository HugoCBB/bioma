package handler

import (
	"context"
	"net/http"

	"github.com/bioma/internal/agent"
	"github.com/bioma/internal/domain"
	"github.com/bioma/internal/repository"
	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	repo *repository.Repository
}

func NewChatHandlerSetup(repo *repository.Repository) *ChatHandler {
	return &ChatHandler{repo: repo}
}

func (r *ChatHandler) handlerChat(c *gin.Context) {
	var payload domain.ChatRequest

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id e message são obrigatórios"})
		return
	}

	token, err := r.repo.GetToken(c, payload.UserId)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":     "usuário não autenticado",
			"login_url": "/auth/login?user_id=" + payload.UserId,
		})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "access_token", token.AccessToken)

	result, err := agent.RunAgent(ctx, payload.Message, token.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": result})
}
