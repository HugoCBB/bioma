package handler

import (
	"context"
	"net/http"

	"github.com/bioma/internal/config"
	"github.com/bioma/internal/repository"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	repo *repository.AuthRepository
}

func NewAuthHandlerSetup(repo *repository.AuthRepository) *AuthHandler {
	return &AuthHandler{repo: repo}
}

var cfg = config.NewGoogleClientSetup()

func (a *AuthHandler) handleLogin(c *gin.Context) {
	userId := c.Query("user_id")
	url := cfg.AuthCodeURL(userId, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusFound, url)
}

func (a *AuthHandler) handleCallback(c *gin.Context) {
	code := c.Query("code")
	userId := c.Query("state")

	token, err := cfg.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := a.repo.SaveToken(c, userId, token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Login realizado com sucesso!",
		"user_id":      userId,
		"access_token": token.AccessToken,
	})
}
