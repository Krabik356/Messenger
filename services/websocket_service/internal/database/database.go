package database

import (
	"context"
	"messenger/services/websocket_service/internal/models"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	pool *pgxpool.Pool
}

func NewDatabase() *Database {
	pool, err := pgxpool.New(context.Background(), "postgres://postgres:111222035@localhost:5432/check")
	if err != nil {
		panic(err)
	}

	return &Database{
		pool: pool,
	}
}

func (db *Database) Close() {
	db.pool.Close()
}

func (db *Database) NewChat(ctx context.Context, chatId int, userId ...int) error {
	ctxTime, stop := context.WithTimeout(ctx, 2*time.Second)
	defer stop()

	tx, err := db.pool.Begin(ctxTime)
	if err != nil {
		return models.ServersError
	}
	defer tx.Rollback(ctxTime)

	for _, id := range userId {
		if _, err := tx.Exec(ctxTime, "INSERT INTO chat_users(chat_id, user_id) VALUES($1, $2)", chatId, id); err != nil {
			return models.ServersError
		}
	}

	return tx.Commit(ctxTime)
}
