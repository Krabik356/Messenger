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

func NewDatabase(addr string) *Database {
	pool, err := pgxpool.New(context.Background(), addr)
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

func (db *Database) AddNewUser(ctx context.Context, id int, name, email string) error {
	ctxTime, stop := context.WithTimeout(ctx, 2*time.Second)
	defer stop()

	if _, err := db.pool.Exec(ctxTime, "INSERT INTO users(id, name, email) VALUES($1, $2, $3)", id, name, email); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
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

func (db *Database) CreateChat(ctx context.Context, creatorId, anotherId int, chatName string) error {
	ctxTime, stop := context.WithTimeout(ctx, 2*time.Second)
	defer stop()

	var existsChatId int
	err := db.pool.QueryRow(ctxTime, "SELECT chat_id FROM chat_users WHERE user_id IN ($1, $2) GROUP BY chat_id HAVING COUNT(*) = 2", creatorId, anotherId).Scan(&existsChatId)
	if err == nil {
		return models.ChatAlreadyExists
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return models.ServersError
	}

	ctxTx, stop := context.WithTimeout(ctx, 2*time.Second)
	defer stop()

	tx, err := db.pool.Begin(ctxTx)
	if err != nil {
		return models.ServersError
	}
	defer tx.Rollback(ctxTx)

	var chatId int
	err = tx.QueryRow(ctxTx, "INSERT INTO chats(name) VALUES($1) RETURNING id", chatName).Scan(&chatId)
	if err != nil {
		return models.ServersError
	}
	if _, err := tx.Exec(ctxTx, "INSERT INTO chat_users(chat_id, user_id) VALUES($1, $2)", chatId, creatorId); err != nil {
		return models.ServersError
	}
	if _, err := tx.Exec(ctxTx, "INSERT INTO chat_users(chat_id, user_id) VALUES($1, $2)", chatId, anotherId); err != nil {
		return models.ServersError
	}

	return tx.Commit(ctxTx)
}

func (db *Database) SendMessage(ctx context.Context, chatId, userId int, message string) (int, error) {
	ctxTime, stop := context.WithTimeout(ctx, 2*time.Second)
	defer stop()

	var id int
	if err := db.pool.QueryRow(ctxTime, "INSERT INTO chat_messages(chat_id, user_id, msg) SELECT $1, $2, $3 WHERE EXISTS(SELECT TRUE FROM chat_users WHERE chat_id=$1 AND user_id=$2) RETURNING id", chatId, userId, message).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, models.NoUserInChat
		}
		return 0, models.ServersError
	}
	return id, nil
}
