package adapter

import (
	"encoding/json"
	"fmt"
	"strings"

	"belajar-golang-unit-test/config"
	"belajar-golang-unit-test/internal/core/domain/entity"
	"belajar-golang-unit-test/internal/core/service"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type MiddlewareAdapterInterface interface {
	CheckToken() fiber.Handler
}

type middlewareAdapter struct {
	cfg        *config.Config
	jwtService service.JwtServiceInterface
	redis      *redis.Client
}

func NewMiddlewareAdapter(cfg *config.Config, jwtService service.JwtServiceInterface, redisClient *redis.Client) MiddlewareAdapterInterface {
	return &middlewareAdapter{
		cfg:        cfg,
		jwtService: jwtService,
		redis:      redisClient,
	}
}

func (m *middlewareAdapter) CheckToken() fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing or invalid authorization header")
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid bearer format")
		}

		claims, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			log.Warn().
				Err(err).
				Str("source", "adapter.middlewareAdapter.CheckToken").
				Msg("invalid jwt token")
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}

		// Verifikasi sesi di Redis jika client redis tersedia
		if m.redis != nil {
			sessionKey := fmt.Sprintf("session:%s", tokenString)
			sessionVal, err := m.redis.Get(c.Context(), sessionKey).Result()
			if err != nil || sessionVal == "" {
				log.Warn().
					Str("source", "adapter.middlewareAdapter.CheckToken").
					Msg("session not found or expired in redis")
				return fiber.NewError(fiber.StatusUnauthorized, "session expired or invalid")
			}

			var jwtUserData entity.JwtUserData
			if err := json.Unmarshal([]byte(sessionVal), &jwtUserData); err == nil {
				c.Locals("user", jwtUserData)
			}
		} else {
			c.Locals("user", entity.JwtUserData{
				ID:    claims.UserID,
				Email: claims.Email,
				Name:  claims.Name,
				Role:  claims.Role,
			})
		}

		return c.Next()
	}
}
