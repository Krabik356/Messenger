package token

import (
	"chat_manager_service/models"
	"errors"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

type Token struct {
}

func NewToken() *Token {
	return &Token{}
}

func (tok *Token) IsValidToken(tokenString string) (int, bool, error) {
	t, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) { return []byte("kraby_secret"), nil })
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, false, models.ExpiredToken
		}
		return 0, false, models.ServersError
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
