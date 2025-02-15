package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var users = map[string]string{
	"user": "password",
}

func Authenticate(username, password string) (string, error) {
	if pwd, ok := users[username]; ok && pwd == password {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": username,
			"exp":      time.Now().Add(24 * time.Hour).Unix(),
		})
		return token.SignedString([]byte("secret-key"))
	}
	return "", errors.New("invalid credentials")
}
