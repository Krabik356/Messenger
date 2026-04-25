package token

import (
	"Messenger/internal/models"
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(ctx context.Context, email, tokenType string) (string, error) {
	var workTime time.Time
	if tokenType == "access" {
		workTime = time.Now().Add(15 * time.Minute)
	} else if tokenType == "refresh" {
		workTime = time.Now().Add(24 * 7 * time.Hour)
	} else {
		return "", models.InvalidTokenType
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"type":  tokenType,
		"exp":   workTime.Unix(),
	})

	strToken, err := token.SignedString([]byte("krabs_secret"))
	if err != nil {
		return "", models.ServersError
	}

	return strToken, nil
}

func IsValidToken(strToken string) (bool, string, error) {
	token, err := jwt.Parse(strToken, func(token *jwt.Token) (interface{}, error) { return []byte("krabs_secret"), nil })
	if err != nil {
		return false, "", models.ServersError
	}
	if token.Method != jwt.SigningMethodHS256 {
		return false, "", models.InvalidToken
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false, "", models.ServersError
	}
	return token.Valid, claims["email"].(string), nil
}
