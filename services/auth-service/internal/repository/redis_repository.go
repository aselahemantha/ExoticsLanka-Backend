package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/exoticsLanka/auth-service/internal/domain"
	"github.com/google/uuid"
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
	pipe := r.client.Pipeline()
	pipe.Set(ctx, key, data, time.Until(session.ExpiresAt))

	// Add token to user's session set
	userKey := fmt.Sprintf("user_sessions:%s", session.UserID.String())
	pipe.SAdd(ctx, userKey, session.Token)
	pipe.Expire(ctx, userKey, time.Until(session.ExpiresAt)) // Refresh expiry of set

	_, err = pipe.Exec(ctx)
	return err
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
	// We need to resolve userID to remove from set, but that's expensive if we don't have the session.
	// Alternative: Get session first, then delete both.
	// Or: Leave it as orphaned in the set (it will expire eventually or handle it on access).
	// For consistence, let's try to get it.

	session, err := r.GetByToken(ctx, token)
	if err != nil {
		return err // Or ignore if nil?
	}
	if session == nil {
		return nil // Already deleted
	}

	key := fmt.Sprintf("session:%s", token)
	userKey := fmt.Sprintf("user_sessions:%s", session.UserID.String())

	pipe := r.client.Pipeline()
	pipe.Del(ctx, key)
	pipe.SRem(ctx, userKey, token)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *redisSessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	userKey := fmt.Sprintf("user_sessions:%s", userID.String())

	// Get all tokens
	tokens, err := r.client.SMembers(ctx, userKey).Result()
	if err != nil {
		return err
	}

	if len(tokens) == 0 {
		return nil
	}

	pipe := r.client.Pipeline()
	// Delete all session keys
	for _, token := range tokens {
		pipe.Del(ctx, fmt.Sprintf("session:%s", token))
	}
	// Delete the set itself
	pipe.Del(ctx, userKey)

	_, err = pipe.Exec(ctx)
	return err
}
