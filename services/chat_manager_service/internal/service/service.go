package service

import (
	"chat_manager_service/internal/database"
	"chat_manager_service/internal/token"
	"context"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Service struct {
	database *database.Database
	token    *token.Token
}

func NewService(database *database.Database, token *token.Token, cons *kafka.Consumer) *Service {
	return &Service{
		database: database,
		token:    token,
	}
}

func (s *Service) Close() {
	s.database.Close()
}

func (s *Service) CreateChat(ctx context.Context, creatorId, anotherId int, chatName string) error {
	return s.database.CreateChat(ctx, creatorId, anotherId, chatName)
}

func (s *Service) IsValidToken(tokenString string) (int, bool, error) {
	return s.token.IsValidToken(tokenString)
}

func (s *Service) AddNewUser(ctx context.Context, id int, name, email string) error {
	if err := s.database.AddNewUser(ctx, id, email); err != nil {
		return err
	}
	return nil
}
