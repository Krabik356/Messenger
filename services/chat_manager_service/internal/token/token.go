package token

import (
	"chat_manager_service/internal/models"
	"errors"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

type Token struct {
	secret string
}

func NewToken(secret string) *Token {
	return &Token{
		secret: secret,
	}
}

func (tok *Token) IsValidToken(tokenString string) (int, bool, error) {
	t, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) { return []byte(tok.secret), nil })
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, false, models.ExpiredToken
		}
		return 0, false, models.ServersError
	}

	if t.Method != jwt.SigningMethodHS256 {
		return 0, false, models.InvalidToken
	}

	if !t.Valid {
		return 0, false, nil
	}
	idString, ok := t.Claims.(jwt.MapClaims)["userID"].(string)
	if !ok {
		return 0, false, models.InvalidToken
	}

	id, err := strconv.Atoi(idString)
	if err != nil {
		return 0, false, models.InvalidToken
	}

	return id, true, nil
}
