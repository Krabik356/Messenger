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
	if _, err := db.pool.Exec(ctxTime, "INSERT INTO users(name, email, password) VALUES($1, $2, $3)", name, email, password); err != nil {
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
	return nil
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
