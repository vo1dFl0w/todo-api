package model

import "testing"

func TestUser(t *testing.T) *User {
	return &User{
		ID: 1,
		Email: "user@example.org",
		Password: "password",
		RefreshToken: "very-secret-key",
	}
}