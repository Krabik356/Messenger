package database

import (
	"Messenger/internal/models"
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

func (db *Database) Register(ctx context.Context, name, email, password string) error {
	ctxTime, stop := context.WithTimeout(ctx, 2*time.Second)
	defer stop()

	tx, err := db.pool.Begin(ctxTime)
	if err != nil {
		return models.ServersError
	}
	defer tx.Rollback(ctxTime)

	var userId int
	if err := tx.QueryRow(ctxTime, "INSERT INTO users(name, email, password) VALUES($1, $2, $3) RETURNING id", name, email, password).Scan(&userId); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return models.AlreadyExists
			case "23502":
				if pgErr.ColumnName == "name" {
					return models.NullName
				} else if pgErr.ColumnName == "email" {
					return models.NullEmail
				}
				return models.NullPassword
			case "23514":
				if pgErr.ColumnName == "name" {
					return models.InvalidName
				} else if pgErr.ColumnName == "email" {
					return models.InvalidEmail
				}
				return models.InvalidPassword
			default:
				return models.ServersError
			}
		}
		return models.ServersError
	}

	if _, err := tx.Exec(ctxTime, "INSERT INTO outbox(user_id, name, email, topic, status) VALUES($1, $2, $3, $4, $5)", userId, name, email, "new_user", "waiting"); err != nil {
		return models.ServersError
	}

	return tx.Commit(ctxTime)
}

func (db *Database) Login(ctx context.Context, email string) (string, error) {
	var password string

	ctxTime, stop := context.WithTimeout(ctx, 2*time.Second)
	defer stop()

	if err := db.pool.QueryRow(ctxTime, "SELECT password FROM users WHERE email=$1", email).Scan(&password); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", models.UnknownUser
		}
		return "", models.ServersError
	}

	return password, nil
}

func (db *Database) GetFromOutbox(ctx context.Context, num int) ([]models.RegForKafka, error) {
	ctxTime, stop := context.WithTimeout(ctx, 2*time.Second)
	defer stop()

	data, err := db.pool.Query(ctxTime, "SELECT id, name, email FROM outbox WHERE status=$1 LIMIT $2", "waiting", num)
	if err != nil {
		return nil, models.ServersError
	}
	defer data.Close()

	var result []models.RegForKafka

	for data.Next() {
		var id int
		var name string
		var email string

		if err := data.Scan(&id, &name, &email); err != nil {
			continue
		}

		result = append(result, models.RegForKafka{id, name, email})
	}

	return result, nil
}

func (db *Database) CommitOutboxByUserId(ctx context.Context, id int) error {
	ctxTime, stop := context.WithTimeout(ctx, 2*time.Second)
	defer stop()

	if _, err := db.pool.Exec(ctxTime, "UPDATE outbox SET status=$1 WHERE user_id=$2", "commited", id); err != nil {
		return models.ServersError
	}
	return nil
}
