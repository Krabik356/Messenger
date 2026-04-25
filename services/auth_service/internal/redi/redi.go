package redi

import (
	"Messenger/internal/models"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	rdb *redis.Client
	ctx context.Context
}

func NewRedis(ctx context.Context) *Redis {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	return &Redis{
		rdb: rdb,
		ctx: ctx,
	}
}

func (r *Redis) Close() error {
	return r.rdb.Close()
}

func (r *Redis) SaveRefreshToken(email, token string) error {
	ctxTime, stop := context.WithTimeout(r.ctx, 2*time.Second)
	defer stop()
	if err := r.rdb.Set(ctxTime, fmt.Sprintf("user:%s", email), token, 24*7*time.Hour).Err(); err != nil {
		return models.ServersError
	}
	return nil
}

func (r *Redis) IsValidToken(strToken, email string) (bool, error) {
	ctxTime, stop := context.WithTimeout(r.ctx, 2*time.Second)
	defer stop()
	token, err := r.rdb.Get(ctxTime, fmt.Sprintf("user:%s", email)).Result()
	if err != nil {
		return false, models.ServersError
	}
	return token == strToken, nil
}
