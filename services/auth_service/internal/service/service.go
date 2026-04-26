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
	tok      *token.Token
}

func NewService(database *database.Database, redi *redi.Redis, tok *token.Token) *Service {
	return &Service{
		database: database,
		redi:     redi,
		tok:      tok,
	}
}

func (s *Service) Close() error {
	s.database.Close()
	return s.redi.Close()
}

func (s *Service) GenerateJWTTokens(ctx context.Context, email string) (string, string, error) {
	refreshToken, err := s.tok.GenerateToken(ctx, email, "refresh")
	if err != nil {
		return "", "", err
	}
	accessToken, err := s.tok.GenerateToken(ctx, email, "access")
	if err != nil {
		return "", "", err
	}
	if err := s.redi.SaveRefreshToken(ctx, email, refreshToken); err != nil {
		return "", "", err
	}

	return refreshToken, accessToken, nil
}

func (s *Service) Register(ctx context.Context, name, email, password string) error {
	if len(name) < 3 || len(name) > 10 {
		return models.InvalidName
	} else if len(email) < 5 || len(email) > 20 {
		return models.InvalidEmail
	} else if len(password) < 5 || len(password) > 20 {
		return models.InvalidPassword
	}
	return s.database.Register(ctx, name, email, password)
}

func (s *Service) Login(ctx context.Context, email, password string) (bool, error) {
	return s.database.Login(ctx, email, password)
}

func (s *Service) IsValidToken(ctx context.Context, strToken string) (bool, string, error) {
	isValid, email, err := s.tok.IsValidToken(strToken)
	if err != nil {
		return false, "", err
	}
	if !isValid {
		return false, email, nil
	}

	isValid2, err := s.redi.IsValidToken(ctx, strToken, email)
	if err != nil {
		return false, email, err
	}
	return isValid2, email, nil
}
