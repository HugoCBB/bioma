package handler

import (
	"context"
	"net/http"

	"github.com/bioma/internal/agent"
	"github.com/bioma/internal/domain"
	"github.com/bioma/internal/repository"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type ChatHandler struct {
	repo  *repository.Repository
	cfg   *oauth2.Config
	agent *agent.Agent
}

func NewChatHandlerSetup(repo *repository.Repository, cfg *oauth2.Config) *ChatHandler {
	a, err := agent.New(nil)
	if err != nil {
		panic("falha ao inicializar agente: " + err.Error())
	}
	return &ChatHandler{
		repo:  repo,
		cfg:   cfg,
		agent: a,
	}
}

func (r *ChatHandler) handlerChat(c *gin.Context) {
	var payload domain.ChatRequest

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id e message são obrigatórios"})
		return
	}

	token, err := r.repo.GetToken(c, payload.UserId, r.cfg)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":     "usuário não autenticado",
			"login_url": "/auth/login?user_id=" + payload.UserId,
		})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "access_token", token.AccessToken)

	result, err := r.agent.Run(ctx, payload.Message, token.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": result})
}
