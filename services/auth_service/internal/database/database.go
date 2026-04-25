package database

import (
	"Messenger/internal/models"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	pool *pgxpool.Pool
	ctx  context.Context
}

func NewDatabase(ctx context.Context) *Database {
	pool, err := pgxpool.New(ctx, "postgres://postgres:111222035@localhost:5432/check")
	if err != nil {
		panic(err)
	}
	return &Database{
		pool: pool,
		ctx:  ctx,
	}
}

func (db *Database) Close() {
	db.pool.Close()
}

func (db *Database) Register(name, email, password string) error {
	ctx, stop := context.WithTimeout(db.ctx, 2*time.Second)
	defer stop()
	if _, err := db.pool.Exec(ctx, "INSERT INTO users(name, email, password) VALUES($1, $2, $3)", name, email, password); err != nil {
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

func (db *Database) Login(email, password string) (bool, error) {
	isExists := false

	ctx, stop := context.WithTimeout(db.ctx, 2*time.Second)
	defer stop()

	if err := db.pool.QueryRow(ctx, "SELECT EXISTS(SELECT TRUE FROM users WHERE email=$1 AND password=$2)", email, password).Scan(&isExists); err != nil {
		return false, models.ServersError
	}

	return isExists, nil
}
