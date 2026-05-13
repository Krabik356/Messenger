package database

import (
	"chat_manager_service/internal/models"
	"context"
	"encoding/json"
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

	tx, err := db.pool.Begin(ctxTime)
	if err != nil {
		return 0, models.ServersError
	}
	defer tx.Rollback(ctxTime)

	var id int
	if err := tx.QueryRow(ctxTime, "INSERT INTO chat_messages(chat_id, user_id, msg) SELECT $1, $2, $3 WHERE EXISTS(SELECT TRUE FROM chat_users WHERE chat_id=$1 AND user_id=$2) RETURNING id", chatId, userId, message).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, models.NoUserInChat
		}
		return 0, models.ServersError
	}
	if _, err := tx.Exec(ctxTime, "INSERT INTO OUTBOX(msg_id, user_id, user_name, user_email, msg, status, topic) SELECT $1, $2, users.name, users.email $3, $4, $5 FROM users WHERE users.id=$2", id, userId, message, "waiting", "messages"); err != nil {
		return 0, models.ServersError
	}

	return id, tx.Commit(ctxTime)
}

func (db *Database) GetUsersData(ctx context.Context, id int) (models.GetUsersDataResponse, error) {
	ctxTime, stop := context.WithTimeout(ctx, 2*time.Second)
	defer stop()

	data, err := db.pool.Query(ctxTime, `

		SELECT cu.chat_id, ch.name, JSONB_AGG(DISTINCT JSONB_BUILD_OBJECT('id', u.id,'name', u.name,'email', u.email)) AS users, JSONB_AGG(DISTINCT JSONB_BUILD_OBJECT('id', m.id, 'writer', JSONB_BUILD_OBJECT('id', w.id,'name', w.name,'email', w.email), 'text', m.text)) AS messages
		FROM chat_users AS cu
		JOIN chats AS ch ON cu.chat_id=ch.id
		JOIN chat_users AS cu2 ON cu.chat_id=cu2.chat_id
		JOIN users AS u ON cu2.user_id=u.id
		JOIN chat_messages AS m ON cu.chat_id=m.chat_id
		JOIN users AS w ON w.id=m.user_id
		WHERE cu.user_id=$1
		GROUP BY cu.chat_id, ch.name

	`, id)
	if err != nil {
		return models.GetUsersDataResponse{}, models.ServersError
	}

	defer data.Close()

	var result models.GetUsersDataResponse
	for data.Next() {
		var (
			messageUnmarshalled   models.MessageSlice
			chatUsersUnmarshalled models.UsersSlice
			chatId                int
			chatName              string
			chatUsers             []byte
			messages              []byte
		)
		if err := data.Scan(&chatId, &chatName, &chatUsers, &messages); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return models.GetUsersDataResponse{}, models.EmptyChat
			}
			return models.GetUsersDataResponse{}, models.ServersError
		}

		if err := json.Unmarshal(messages, &messageUnmarshalled); err != nil {
			return models.GetUsersDataResponse{}, models.ServersError
		}

		if err := json.Unmarshal(chatUsers, &chatUsersUnmarshalled); err != nil {
			return models.GetUsersDataResponse{}, models.ServersError
		}

		result.Chats = append(result.Chats, models.Chat{
			Id:       chatId,
			Name:     chatName,
			Users:    chatUsersUnmarshalled,
			Messages: messageUnmarshalled,
		})
	}

	return result, nil

}

func (db *Database) GetChatsMessages(ctx context.Context, chatId, offset int) ([]models.Message, error) {
	ctxTime, stop := context.WithTimeout(ctx, 2*time.Second)
	defer stop()

	data, err := db.pool.Query(ctxTime, `

		SELECT m.id, m.msg, JSONB_BUILD_OBJECT('id', u.id,'name', u.name,'email', u.email) 
		FROM chat_messages AS m
		JOIN users AS u ON u.id=m.user_id
		WHERE m.chat_id=$1
		OFFSET $2
		LIMIT 50
		ORDER BY m.id DESC

	`, chatId, offset)

	if err != nil {
		return nil, models.ServersError
	}

	defer data.Close()

	var result []models.Message

	for data.Next() {
		var (
			msgId   int
			msgText string
			rawUser []byte
			user    models.User
		)

		if err := data.Scan(&msgId, &msgText, &rawUser); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, models.NoOldMessages
			}
			return nil, models.ServersError
		}

		if err := json.Unmarshal(rawUser, &user); err != nil {
			return nil, models.ServersError
		}

		result = append(result, models.Message{
			Id:   msgId,
			User: user,
			Text: msgText,
		})
	}

	return result, nil
}

func (db *Database) GetFromOutbox(ctx context.Context) ([]models.Message, error) {
	ctxTime, stop := context.WithTimeout(ctx, 2*time.Second)
	defer stop()

	data, err := db.pool.Query(ctxTime, "SELECT * FROM outbox WHERE status=$1 AND topic=$2 LIMIT 50", "waiting", "messages")
	if err != nil {
		return nil, models.ServersError
	}
	defer data.Close()

	var result []models.Message
	for data.Next() {
		var (
			msgId     int
			userId    int
			userName  string
			userEmail string
			msg       string
		)

		if err := data.Scan(&msgId, &userId, &userName, &userEmail, &msg); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, nil
			}
			return nil, models.ServersError
		}

		result = append(result, models.Message{msgId, models.User{userId, userName, userEmail}, msg})
	}

	return result, nil
}

func (db *Database) CommitMessage(ctx context.Context, msgId int) error {
	ctxTime, stop := context.WithTimeout(ctx, 2*time.Second)
	defer stop()

	rowsAffected, err := db.pool.Exec(ctxTime, "UPDATE outbox SET status=$1 WHERE msg_id=$2", "sended", msgId)
	if err != nil {
		return models.ServersError
	}
	if rowsAffected.RowsAffected() == 0 {
		return models.ServersError
	}

	return nil
}
