package service

import (
	"chat_manager_service/internal/database"
	"chat_manager_service/internal/models"
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

func (s *Service) CreateChat(ctx context.Context, creatorId, anotherId int, chatName string) error {
	return s.database.CreateChat(ctx, creatorId, anotherId, chatName)
}

func (s *Service) IsValidToken(tokenString string) (int, bool, error) {
	return s.token.IsValidToken(tokenString)
}

func (s *Service) AddNewUser(ctx context.Context, id int, name, email string) error {
	return s.database.AddNewUser(ctx, id, name, email)
}

func (s *Service) SendMessage(ctx context.Context, chatId, userId int, message string) (int, error) {
	return s.database.SendMessage(ctx, chatId, userId, message)
}

func (s *Service) GetUsersData(ctx context.Context, id int) (models.GetUsersDataResponse, error) {
	return s.database.GetUsersData(ctx, id)
}

func (s *Service) GetChatsMessages(ctx context.Context, chatId, offset int) ([]models.Message, error) {
	return s.database.GetChatsMessages(ctx, chatId, offset)
}

func (s *Service) GetFromOutbox(ctx context.Context) ([]models.Message, error) {
	return s.database.GetFromOutbox(ctx)
}

func (s *Service) CommitMessage(ctx context.Context, msgId int) error {
	return s.database.CommitMessage(ctx, msgId)
}
