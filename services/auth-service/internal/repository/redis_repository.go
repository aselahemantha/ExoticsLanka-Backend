package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/exoticsLanka/auth-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

type redisSessionRepository struct {
	client *redis.Client
}

// NewRedisSessionRepository creates a new session repository
func NewRedisSessionRepository(client *redis.Client) domain.SessionRepository {
	return &redisSessionRepository{client: client}
}

func (r *redisSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("session:%s", session.Token)
	return r.client.Set(ctx, key, data, time.Until(session.ExpiresAt)).Err()
}

func (r *redisSessionRepository) GetByToken(ctx context.Context, token string) (*domain.Session, error) {
	key := fmt.Sprintf("session:%s", token)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var session domain.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *redisSessionRepository) Delete(ctx context.Context, token string) error {
	key := fmt.Sprintf("session:%s", token)
	return r.client.Del(ctx, key).Err()
}
