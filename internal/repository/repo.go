package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
)

type Repository struct {
	rdb *redis.Client
}

func NewRepository(rdb *redis.Client) *Repository {
	return &Repository{rdb: rdb}
}

const tokenKeyPrefix = "user:token:"

func tokenKey(userId string) string {
	return fmt.Sprintf("%s%s", tokenKeyPrefix, userId)
}

func (a *Repository) SaveToken(ctx context.Context, userId string, token *oauth2.Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("erro ao serializar token: %w", err)
	}

	ttl := time.Until(token.Expiry)
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}

	if err := a.rdb.Set(ctx, tokenKey(userId), data, ttl).Err(); err != nil {
		return fmt.Errorf("erro ao salvar token no redis: %w", err)
	}

	return nil
}

func (a *Repository) GetToken(ctx context.Context, userId string, cfg *oauth2.Config) (*oauth2.Token, error) {
	data, err := a.rdb.Get(ctx, tokenKey(userId)).Bytes()
	if err != nil {
		return nil, fmt.Errorf("token não encontrado para user %s: %w", userId, err)
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("erro ao desserializar token: %w", err)
	}

	if token.Expiry.Before(time.Now()) {
		tokenSource := cfg.TokenSource(ctx, &token)
		newToken, err := tokenSource.Token()
		if err != nil {
			return nil, fmt.Errorf("erro ao renovar token: %w", err)
		}

		if err := a.SaveToken(ctx, userId, newToken); err != nil {
			return nil, err
		}
		return newToken, nil

	}
	return &token, nil
}
