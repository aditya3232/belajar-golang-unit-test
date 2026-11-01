package service_test

import (
	"testing"
	"time"

	"belajar-golang-unit-test/config"
	"belajar-golang-unit-test/internal/core/service"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestConfig() *config.Config {
	return &config.Config{
		App: config.App{
			JwtSecretKey: "test-secret-key-12345",
			JwtIssuer:    "test-issuer",
		},
	}
}

func TestJwtService_GenerateAndValidateToken(t *testing.T) {
	cfg := newTestConfig()
	jwtSvc := service.NewJwtService(cfg)

	t.Run("sukses generate dan validasi token", func(t *testing.T) {
		token, err := jwtSvc.GenerateToken(1, "adit@example.com", "Adit", "user")
		require.NoError(t, err)
		require.NotEmpty(t, token)

		claims, err := jwtSvc.ValidateToken(token)
		require.NoError(t, err)
		require.NotNil(t, claims)

		assert.Equal(t, int64(1), claims.UserID)
		assert.Equal(t, "adit@example.com", claims.Email)
		assert.Equal(t, "Adit", claims.Name)
		assert.Equal(t, "user", claims.Role)
	})

	t.Run("gagal validasi token rusak (malformed)", func(t *testing.T) {
		claims, err := jwtSvc.ValidateToken("ini.bukan.tokenjwtvalid")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("gagal validasi token dengan secret berbeda", func(t *testing.T) {
		otherCfg := &config.Config{
			App: config.App{
				JwtSecretKey: "secret-berbeda-lainnya",
				JwtIssuer:    "test-issuer",
			},
		}
		otherJwtSvc := service.NewJwtService(otherCfg)

		token, err := otherJwtSvc.GenerateToken(2, "budi@example.com", "Budi", "user")
		require.NoError(t, err)

		// Verifikasi dengan jwtSvc pertama (secret beda)
		claims, err := jwtSvc.ValidateToken(token)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("gagal validasi token yang sudah kedaluwarsa (expired)", func(t *testing.T) {
		// Buat token yang sudah expired secara manual
		expiredClaims := service.JwtCustomClaims{
			UserID: 3,
			Email:  "expired@example.com",
			Name:   "Expired User",
			Role:   "user",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    cfg.App.JwtIssuer,
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // 1 jam yang lalu
			},
		}
		expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
		signedExpiredToken, err := expiredToken.SignedString([]byte(cfg.App.JwtSecretKey))
		require.NoError(t, err)

		claims, err := jwtSvc.ValidateToken(signedExpiredToken)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}
