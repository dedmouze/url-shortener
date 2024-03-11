package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Token struct {
	UID        int64
	Email      string
	Expiration time.Time
	Level      int8
}

func Parse(raw, appSecret string) (*Token, error) {
	const op = "lib.jwt.Parse"

	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(raw, &claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	token := &Token{
		UID:        int64(claims["uid"].(float64)),
		Email:      claims["email"].(string),
		Expiration: time.Unix(int64(claims["exp"].(float64)), 0),
		Level:      int8(claims["level"].(float64)),
	}

	return token, nil
}
