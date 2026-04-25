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
