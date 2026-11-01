package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"belajar-golang-unit-test/config"
	"belajar-golang-unit-test/internal/adapter/handler"
	"belajar-golang-unit-test/internal/adapter/handler/request"
	"belajar-golang-unit-test/internal/adapter/handler/response"
	"belajar-golang-unit-test/internal/adapter/repository"
	"belajar-golang-unit-test/internal/core/domain/entity"
	"belajar-golang-unit-test/internal/core/service"

	"github.com/go-redis/redismock/v9"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUserService mengimplementasikan service.UserServiceInterface untuk pengujian handler
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) SignIn(ctx context.Context, email, password string) (*entity.UserEntity, string, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).(*entity.UserEntity), args.String(1), args.Error(2)
}

func (m *MockUserService) SignUp(ctx context.Context, req entity.UserEntity) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockUserService) GetProfile(ctx context.Context, userID int64) (*entity.UserEntity, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserEntity), args.Error(1)
}

func (m *MockUserService) UpdateProfile(ctx context.Context, req entity.UserEntity) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockUserService) DeleteUser(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserService) GetAllUsers(ctx context.Context, query entity.QueryStringUser) ([]entity.UserEntity, int64, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]entity.UserEntity), args.Get(1).(int64), args.Error(2)
}

func setupHandlerTestApp(t *testing.T) (*fiber.App, *MockUserService, service.JwtServiceInterface, redismock.ClientMock) {
	cfg := &config.Config{
		App: config.App{
			JwtSecretKey: "secret-key-handler-test",
			JwtIssuer:    "test-issuer",
		},
	}

	mockUserSvc := new(MockUserService)
	jwtSvc := service.NewJwtService(cfg)
	redisClient, mockRedis := redismock.NewClientMock()

	app := cfg.NewFiber()
	handler.NewUserHandler(app, mockUserSvc, cfg, jwtSvc, redisClient)

	t.Cleanup(func() {
		mockUserSvc.AssertExpectations(t)
	})

	return app, mockUserSvc, jwtSvc, mockRedis
}

func TestUserHandler_SignUp(t *testing.T) {
	t.Run("sukses register user baru (201 Created)", func(t *testing.T) {
		app, mockSvc, _, _ := setupHandlerTestApp(t)

		reqBody := request.SignUpRequest{
			Name:     "Aditya",
			Email:    "adit@example.com",
			Password: "securePassword123",
			Phone:    "081234567890",
			Address:  "Jakarta",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockSvc.On("SignUp", mock.Anything, mock.MatchedBy(func(u entity.UserEntity) bool {
			return u.Email == "adit@example.com" && u.Name == "Aditya"
		})).Return(nil)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(bodyBytes))
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(httpReq)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

		respBytes, _ := io.ReadAll(resp.Body)
		var defResp response.DefaultResponse
		_ = json.Unmarshal(respBytes, &defResp)
		assert.Equal(t, "user registered successfully", defResp.Message)
	})

	t.Run("validasi gagal karena password terlalu pendek (400 Bad Request)", func(t *testing.T) {
		app, _, _, _ := setupHandlerTestApp(t)

		reqBody := request.SignUpRequest{
			Name:     "Aditya",
			Email:    "adit@example.com",
			Password: "123", // min 6
		}
		bodyBytes, _ := json.Marshal(reqBody)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(bodyBytes))
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(httpReq)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("email sudah terdaftar (409 Conflict)", func(t *testing.T) {
		app, mockSvc, _, _ := setupHandlerTestApp(t)

		reqBody := request.SignUpRequest{
			Name:     "Existing User",
			Email:    "exist@example.com",
			Password: "securePassword123",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockSvc.On("SignUp", mock.Anything, mock.Anything).Return(service.ErrUserAlreadyExists)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(bodyBytes))
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(httpReq)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusConflict, resp.StatusCode)
	})
}

func TestUserHandler_SignIn(t *testing.T) {
	t.Run("sukses login (200 OK)", func(t *testing.T) {
		app, mockSvc, _, _ := setupHandlerTestApp(t)

		reqBody := request.SignInRequest{
			Email:    "adit@example.com",
			Password: "passwordRahasia",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		expectedUser := &entity.UserEntity{
			ID:    1,
			Name:  "Adit",
			Email: "adit@example.com",
			Role:  "user",
		}
		mockSvc.On("SignIn", mock.Anything, "adit@example.com", "passwordRahasia").
			Return(expectedUser, "valid-token-jwt-123", nil)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signin", bytes.NewReader(bodyBytes))
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(httpReq)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		respBytes, _ := io.ReadAll(resp.Body)
		var defResp response.DefaultResponse
		_ = json.Unmarshal(respBytes, &defResp)
		assert.Equal(t, "signin successful", defResp.Message)
	})

	t.Run("gagal login kredensial salah (401 Unauthorized)", func(t *testing.T) {
		app, mockSvc, _, _ := setupHandlerTestApp(t)

		reqBody := request.SignInRequest{
			Email:    "adit@example.com",
			Password: "passwordSalah",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockSvc.On("SignIn", mock.Anything, "adit@example.com", "passwordSalah").
			Return(nil, "", service.ErrInvalidPassword)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signin", bytes.NewReader(bodyBytes))
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(httpReq)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})
}

func TestUserHandler_GetProfile(t *testing.T) {
	t.Run("sukses ambil profil dengan token & redis session valid", func(t *testing.T) {
		app, mockSvc, jwtSvc, mockRedis := setupHandlerTestApp(t)

		token, err := jwtSvc.GenerateToken(5, "profile@example.com", "Profile User", "user")
		require.NoError(t, err)

		sessionData := entity.JwtUserData{ID: 5, Email: "profile@example.com", Name: "Profile User", Role: "user"}
		sessionBytes, _ := json.Marshal(sessionData)

		mockRedis.ExpectGet("session:" + token).SetVal(string(sessionBytes))

		expectedProfile := &entity.UserEntity{
			ID:      5,
			Name:    "Profile User",
			Email:   "profile@example.com",
			Address: "Yogyakarta",
		}
		mockSvc.On("GetProfile", mock.Anything, int64(5)).Return(expectedProfile, nil)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/users/profile", nil)
		httpReq.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(httpReq)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("profil tidak ditemukan di database (404 Not Found)", func(t *testing.T) {
		app, mockSvc, jwtSvc, mockRedis := setupHandlerTestApp(t)

		token, err := jwtSvc.GenerateToken(99, "notfound@example.com", "Not Found", "user")
		require.NoError(t, err)

		sessionData := entity.JwtUserData{ID: 99, Email: "notfound@example.com", Name: "Not Found"}
		sessionBytes, _ := json.Marshal(sessionData)

		mockRedis.ExpectGet("session:" + token).SetVal(string(sessionBytes))

		mockSvc.On("GetProfile", mock.Anything, int64(99)).Return(nil, repository.ErrUserNotFound)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/users/profile", nil)
		httpReq.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(httpReq)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

func TestUserHandler_UpdateProfile(t *testing.T) {
	t.Run("sukses update profil (200 OK)", func(t *testing.T) {
		app, mockSvc, jwtSvc, mockRedis := setupHandlerTestApp(t)

		token, err := jwtSvc.GenerateToken(5, "profile@example.com", "Profile User", "user")
		require.NoError(t, err)

		sessionData := entity.JwtUserData{ID: 5, Email: "profile@example.com", Name: "Profile User", Role: "user"}
		sessionBytes, _ := json.Marshal(sessionData)
		mockRedis.ExpectGet("session:" + token).SetVal(string(sessionBytes))

		reqBody := request.UpdateProfileRequest{
			Name:    "Nama Baru",
			Address: "Alamat Baru",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockSvc.On("UpdateProfile", mock.Anything, mock.MatchedBy(func(u entity.UserEntity) bool {
			return u.ID == 5 && u.Name == "Nama Baru"
		})).Return(nil)

		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/users/profile", bytes.NewReader(bodyBytes))
		httpReq.Header.Set("Authorization", "Bearer "+token)
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(httpReq)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}

func TestUserHandler_DeleteUser(t *testing.T) {
	t.Run("sukses delete user (200 OK)", func(t *testing.T) {
		app, mockSvc, jwtSvc, mockRedis := setupHandlerTestApp(t)

		token, _ := jwtSvc.GenerateToken(1, "admin@example.com", "Admin", "admin")
		sessionData, _ := json.Marshal(entity.JwtUserData{ID: 1, Email: "admin@example.com", Role: "admin"})
		mockRedis.ExpectGet("session:" + token).SetVal(string(sessionData))

		mockSvc.On("DeleteUser", mock.Anything, int64(42)).Return(nil)

		httpReq := httptest.NewRequest(http.MethodDelete, "/api/v1/users/42", nil)
		httpReq.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(httpReq)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("parameter user id tidak valid (400 Bad Request)", func(t *testing.T) {
		app, _, jwtSvc, mockRedis := setupHandlerTestApp(t)

		token, _ := jwtSvc.GenerateToken(1, "admin@example.com", "Admin", "admin")
		sessionData, _ := json.Marshal(entity.JwtUserData{ID: 1, Email: "admin@example.com", Role: "admin"})
		mockRedis.ExpectGet("session:" + token).SetVal(string(sessionData))

		httpReq := httptest.NewRequest(http.MethodDelete, "/api/v1/users/bukan-angka", nil)
		httpReq.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(httpReq)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}
