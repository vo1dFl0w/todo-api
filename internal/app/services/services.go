package services

type TokenService interface {
	GenerateAccessToken(id int) (string, error)
	GenerateRefreshToken() (string, error)
}