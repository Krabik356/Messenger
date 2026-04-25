package service

import (
	"Messenger/internal/database"
	"Messenger/internal/models"
	"Messenger/internal/redi"
	"Messenger/internal/token"
	"context"
)

type Service struct {
	database *database.Database
	redi     *redi.Redis
	ctx      context.Context
}

func NewService(ctx context.Context, database *database.Database, redi *redi.Redis) *Service {
	return &Service{
		database: database,
		redi:     redi,
		ctx:      ctx,
	}
}

func (s *Service) Close() error {
	s.database.Close()
	return s.redi.Close()
}

func (s *Service) GenerateJWTTokens(email string) (string, string, error) {
	refreshToken, err := token.GenerateToken(s.ctx, email, "refresh")
	if err != nil {
		return "", "", err
	}
	accessToken, err := token.GenerateToken(s.ctx, email, "access")
	if err != nil {
		return "", "", err
	}
	if err := s.redi.SaveRefreshToken(email, refreshToken); err != nil {
		return "", "", err
	}

	return refreshToken, accessToken, nil
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

func (s *Service) Login(email, password string) (bool, error) {
	return s.database.Login(email, password)
}

func (s *Service) IsValidToken(strToken string) (bool, string, error) {
	isValid, email, err := token.IsValidToken(strToken)
	if err != nil {
		return false, "", err
	}
	if !isValid {
		return false, email, nil
	}

	isValid2, err := s.redi.IsValidToken(strToken, email)
	if err != nil {
		return false, email, err
	}
	return isValid2, email, nil
}
