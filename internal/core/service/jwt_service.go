package service

import (
	"errors"
	"time"

	"belajar-golang-unit-test/config"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid or expired token")
)

type jwtService struct {
	secretKey string
	issuer    string
}

type JwtCustomClaims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type JwtServiceInterface interface {
	GenerateToken(userID int64, email, name, role string) (string, error)
	ValidateToken(tokenString string) (*JwtCustomClaims, error)
}

func NewJwtService(cfg *config.Config) JwtServiceInterface {
	return &jwtService{
		secretKey: cfg.App.JwtSecretKey,
		issuer:    cfg.App.JwtIssuer,
	}
}

func (j *jwtService) GenerateToken(userID int64, email, name, role string) (string, error) {
	claims := JwtCustomClaims{
		UserID: userID,
		Email:  email,
		Name:   name,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

func (j *jwtService) ValidateToken(tokenString string) (*JwtCustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JwtCustomClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JwtCustomClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
