package model

import (
	"testing"
	"time"
)

func TestUser(t *testing.T) *User {
	return &User{
		ID: 1,
		Email: "user@example.org",
		Password: "password",
		RefreshToken: "very-secret-key",
		RefreshTokenExpire: time.Now().Add(time.Hour),
	}
}