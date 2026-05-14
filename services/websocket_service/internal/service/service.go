package service

import (
	"context"
	"messenger/services/websocket_service/internal/database"
)

type Service struct {
	database *database.Database
}

func NewService(database *database.Database) *Service {
	return &Service{
		database: database,
	}
}

func (s *Service) NewChat(ctx context.Context, chatId int, userId ...int) error {
	return s.database.NewChat(ctx, chatId, userId...)
}
