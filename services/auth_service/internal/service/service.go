package service

import (
	"Messenger/internal/database"
	"Messenger/internal/models"
	"Messenger/internal/redi"
	"Messenger/internal/token"
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
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
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return models.ServersError
	}
	return s.database.Register(ctx, name, email, string(hashPassword))
}

func (s *Service) Login(ctx context.Context, email, password string) (bool, error) {
	passwordFromDb, err := s.database.Login(ctx, email)
	if err != nil {
		return false, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordFromDb), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, models.ServersError
	}

	return true, nil
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

func (s *Service) GetFromOutbox(ctx context.Context, num int) ([]models.RegForKafka, error) {
	return s.database.GetFromOutbox(ctx, num)
}

func (s *Service) CommitOutboxByUserId(ctx context.Context, id int) error {
	return s.database.CommitOutboxByUserId(ctx, id)
}
