package services

import (
	"RemainsManager/internal/repositories"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"time"
)

type AuthService struct {
	repo      *repositories.AuthRepository
	jwtSecret string
}

func NewAuthService(repo *repositories.AuthRepository, secret string) *AuthService {
	return &AuthService{repo: repo, jwtSecret: secret}
}

func (s *AuthService) Authenticate(username, password string) (string, error) {
	hashedPass, err := s.repo.GetUserByUsername(username)
	if err != nil || hashedPass == "" /*|| hashedPass != localutils.HashSum(username, password)*/ {
		return "", errors.New("Invalid username or password")
	}

	claims := &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
		Issuer:    "RemainManager",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
