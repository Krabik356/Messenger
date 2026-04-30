package service

import (
	"chat_manager_service/internal/database"
	"chat_manager_service/internal/token"
	"context"
)

type Service struct {
	database *database.Database
	token    *token.Token
}

func NewService(database *database.Database, token *token.Token) *Service {
	return &Service{
		database: database,
		token:    token,
	}
}

func (s *Service) Close() {
	s.database.Close()
}

func (s *Service) CreateChat(ctx context.Context, creatorId, anotherId int) error {
	return s.database.CreateChat(ctx, creatorId, anotherId)
}

func (s *Service) IsValidToken(tokenString string) (int, bool, error) {
	return s.token.IsValidToken(tokenString)
}
