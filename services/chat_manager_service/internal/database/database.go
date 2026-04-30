package database

import (
	"chat_manager_service/internal/models"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	pool *pgxpool.Pool
}

func NewDatabase(ctx context.Context) *Database {
	pool, err := pgxpool.New(ctx, "postgres://postgres:111222035@localhost:5432/check")
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

func (db *Database) AddNewUser(ctx context.Context, id int, email string) error {
	ctxTime, stop := context.WithTimeout(ctx, 2*time.Second)
	defer stop()

	if _, err := db.pool.Exec(ctxTime, "INSERT INTO users(id, email) VALUES($1, $2)", id, email); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, pgErr) {
			switch pgErr.Code {
			case "23505":
				return models.AlreadyExists
			default:
				return models.ServersError
			}
		}
		return models.ServersError
	}

	return nil
}

func (db *Database) CreateChat(ctx context.Context, creatorId, anotherId int) error {
	ctxTime, stop := context.WithTimeout(ctx, 2*time.Second)
	defer stop()

	var existsId int
	err := db.pool.QueryRow(ctxTime, "SELECT id FROM chats WHERE users_id=$1", []int{creatorId, anotherId}).Scan(&existsId)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return models.ChatAlreadyExists
		}
	} else if err == nil {
		return models.ChatAlreadyExists
	}

	if _, err := db.pool.Exec(ctxTime, "INSERT INTO chats(users_id) VALUES($1)", []int{creatorId, anotherId}); err != nil {
		return models.ServersError
	}
	return nil
}
