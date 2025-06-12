package auth

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

type tokenClaims struct {
	UserID int `json:"user_id"`
	jwt.StandardClaims
}

type tokenService struct {
	secret []byte
}

func NewTokenService(secret []byte) *tokenService {
	return &tokenService{secret: secret}
}

func (s *tokenService) GenerateAccessToken(id int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		UserID: id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
			IssuedAt: time.Now().Unix(),
		},
	})

	return token.SignedString(s.secret)
}

func (s *tokenService) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", b), nil
}