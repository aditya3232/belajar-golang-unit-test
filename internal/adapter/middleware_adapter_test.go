package adapter_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"belajar-golang-unit-test/config"
	"belajar-golang-unit-test/internal/adapter"
	"belajar-golang-unit-test/internal/core/domain/entity"
	"belajar-golang-unit-test/internal/core/service"

	"github.com/go-redis/redismock/v9"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMiddlewareTestApp(t *testing.T) (*fiber.App, service.JwtServiceInterface, redismock.ClientMock) {
	cfg := &config.Config{
		App: config.App{
			JwtSecretKey: "secret-key-middleware-test",
			JwtIssuer:    "test-issuer",
		},
	}

	jwtSvc := service.NewJwtService(cfg)
	redisClient, mockRedis := redismock.NewClientMock()
	midAdapter := adapter.NewMiddlewareAdapter(cfg, jwtSvc, redisClient)

	app := fiber.New()
	app.Get("/protected", midAdapter.CheckToken(), func(c fiber.Ctx) error {
		user, ok := c.Locals("user").(entity.JwtUserData)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).SendString("user context not found")
		}
		return c.Status(fiber.StatusOK).JSON(user)
	})

	t.Cleanup(func() {
		assert.NoError(t, mockRedis.ExpectationsWereMet())
	})

	return app, jwtSvc, mockRedis
}

func TestMiddlewareAdapter_CheckToken(t *testing.T) {
	t.Run("gagal jika header Authorization tidak ada (401)", func(t *testing.T) {
		app, _, _ := setupMiddlewareTestApp(t)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("gagal jika format token bukan Bearer (401)", func(t *testing.T) {
		app, _, _ := setupMiddlewareTestApp(t)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("gagal jika token JWT tidak valid (401)", func(t *testing.T) {
		app, _, _ := setupMiddlewareTestApp(t)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer token-palsu-ngawur")
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("gagal jika sesi di Redis tidak ditemukan/kedaluwarsa (401)", func(t *testing.T) {
		app, jwtSvc, mockRedis := setupMiddlewareTestApp(t)

		validToken, err := jwtSvc.GenerateToken(1, "adit@example.com", "Adit", "user")
		require.NoError(t, err)

		// Ekspektasi: Redis mengembalikan Nil (sesi sudah kedaluwarsa)
		mockRedis.ExpectGet("session:" + validToken).RedisNil()

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("sukses otentikasi token dan sesi valid (200 OK)", func(t *testing.T) {
		app, jwtSvc, mockRedis := setupMiddlewareTestApp(t)

		validToken, err := jwtSvc.GenerateToken(10, "citra@example.com", "Citra", "admin")
		require.NoError(t, err)

		userData := entity.JwtUserData{
			ID:    10,
			Email: "citra@example.com",
			Name:  "Citra",
			Role:  "admin",
		}
		userBytes, _ := json.Marshal(userData)

		// Ekspektasi: Redis mengembalikan data sesi JSON
		mockRedis.ExpectGet("session:" + validToken).SetVal(string(userBytes))

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		bodyBytes, _ := io.ReadAll(resp.Body)
		var respUser entity.JwtUserData
		_ = json.Unmarshal(bodyBytes, &respUser)
		assert.Equal(t, int64(10), respUser.ID)
		assert.Equal(t, "citra@example.com", respUser.Email)
		assert.Equal(t, "Citra", respUser.Name)
	})
}
