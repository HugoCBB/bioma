package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/bioma/internal/repository"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	repo *repository.Repository
	cfg  *oauth2.Config
}

func NewAuthHandlerSetup(repo *repository.Repository, cfg *oauth2.Config) *AuthHandler {
	return &AuthHandler{
		repo: repo,
		cfg:  cfg,
	}
}

func (a *AuthHandler) handleLogin(c *gin.Context) {
	userId := c.Query("user_id")
	url := a.cfg.AuthCodeURL(userId, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusFound, url)
}

func (a *AuthHandler) handleCallback(c *gin.Context) {
	code := c.Query("code")
	userId := c.Query("state")

	token, err := a.cfg.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("callback recebido - userId: %s, code: %s", userId, code)

	if err := a.repo.SaveToken(c, userId, token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	log.Printf("token salvo com sucesso para userId: %s", userId)
	c.JSON(http.StatusOK, gin.H{
		"message":      "Login realizado com sucesso!",
		"user_id":      userId,
		"access_token": token.AccessToken,
	})
}
