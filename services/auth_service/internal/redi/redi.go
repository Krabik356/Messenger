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
