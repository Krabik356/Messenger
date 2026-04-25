package service

import (
	"RegistrationForMessenger/internal/database"
	"RegistrationForMessenger/internal/models"
	"context"
)

type Service struct {
	database *database.Database
	ctx      context.Context
}

func NewService(ctx context.Context, database *database.Database) *Service {
	return &Service{
		database: database,
		ctx:      ctx,
	}
}

func (s *Service) Register(name, email, password string) error {
	if len(name) < 3 || len(name) > 10 {
		return models.InvalidName
	} else if len(email) < 5 || len(email) > 20 {
		return models.InvalidEmail
	} else if len(password) < 5 || len(password) > 20 {
		return models.InvalidPassword
	}
	return s.database.Register(name, email, password)
}
