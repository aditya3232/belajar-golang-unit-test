package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"belajar-golang-unit-test/config"
	"belajar-golang-unit-test/internal/adapter/repository"
	"belajar-golang-unit-test/internal/core/domain/entity"
	"belajar-golang-unit-test/internal/core/service"
	"belajar-golang-unit-test/internal/core/service/mocks"
	"belajar-golang-unit-test/utils/conv"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupServiceTest(t *testing.T) (*mocks.UserRepositoryMock, *redis.Client, redismock.ClientMock, service.UserServiceInterface, *config.Config) {
	mockRepo := new(mocks.UserRepositoryMock)
	redisClient, mockRedis := redismock.NewClientMock()

	cfg := &config.Config{
		App: config.App{
			JwtSecretKey: "secret-key-test-unit",
			JwtIssuer:    "test-issuer",
		},
	}

	jwtSvc := service.NewJwtService(cfg)
	userSvc := service.NewUserService(mockRepo, jwtSvc, redisClient, cfg)

	t.Cleanup(func() {
		mockRepo.AssertExpectations(t)
		assert.NoError(t, mockRedis.ExpectationsWereMet(), "Ekspektasi Redis mock harus terpenuhi")
	})

	return mockRepo, redisClient, mockRedis, userSvc, cfg
}

func TestUserService_GetProfile_CacheAside(t *testing.T) {
	ctx := context.Background()

	t.Run("skenario 1: CACHE HIT (data profil diambil dari Redis tanpa query database)", func(t *testing.T) {
		mockRepo, _, mockRedis, userSvc, _ := setupServiceTest(t)

		cachedUser := entity.UserEntity{
			ID:      1,
			Name:    "Adit Cached",
			Email:   "adit@example.com",
			Role:    "user",
			Address: "Jakarta",
		}
		cachedBytes, _ := json.Marshal(cachedUser)

		// Ekspektasi: Redis mengembalikan data JSON (Cache Hit)
		mockRedis.ExpectGet("user:profile:1").SetVal(string(cachedBytes))

		// Eksekusi
		result, err := userSvc.GetProfile(ctx, 1)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int64(1), result.ID)
		assert.Equal(t, "Adit Cached", result.Name)

		// Verifikasi krusial: Database repository TIDAK BOLEH dipanggil sama sekali saat Cache Hit!
		mockRepo.AssertNotCalled(t, "GetUserByID", mock.Anything, mock.Anything)
	})

	t.Run("skenario 2: CACHE MISS (data tidak ada di Redis, query database, lalu simpan ke Redis)", func(t *testing.T) {
		mockRepo, _, mockRedis, userSvc, _ := setupServiceTest(t)

		dbUser := &entity.UserEntity{
			ID:      2,
			Name:    "Budi Database",
			Email:   "budi@example.com",
			Role:    "user",
			Address: "Surabaya",
		}

		// 1. Redis Cache Miss: redis.Nil
		mockRedis.ExpectGet("user:profile:2").RedisNil()

		// 2. Query ke Database Repository
		mockRepo.On("GetUserByID", mock.Anything, int64(2)).Return(dbUser, nil)

		// 3. Simpan hasil query ke Redis dengan TTL 15 menit
		mockRedis.Regexp().ExpectSet("user:profile:2", `.+`, 15*time.Minute).SetVal("OK")

		// Eksekusi
		result, err := userSvc.GetProfile(ctx, 2)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int64(2), result.ID)
		assert.Equal(t, "Budi Database", result.Name)
	})

	t.Run("skenario 3: DATA TIDAK ADA DI DATABASE (Cache miss + Repo 404)", func(t *testing.T) {
		mockRepo, _, mockRedis, userSvc, _ := setupServiceTest(t)

		mockRedis.ExpectGet("user:profile:999").RedisNil()
		mockRepo.On("GetUserByID", mock.Anything, int64(999)).Return(nil, repository.ErrUserNotFound)

		result, err := userSvc.GetProfile(ctx, 999)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, repository.ErrUserNotFound, err)
	})
}

func TestUserService_SignIn(t *testing.T) {
	ctx := context.Background()

	t.Run("sukses login dan simpan sesi ke Redis", func(t *testing.T) {
		mockRepo, _, mockRedis, userSvc, _ := setupServiceTest(t)

		password := "rahasia123"
		hashed, _ := conv.HashPassword(password)

		existingUser := &entity.UserEntity{
			ID:       10,
			Name:     "Login User",
			Email:    "login@example.com",
			Password: hashed,
			Role:     "user",
		}

		mockRepo.On("GetUserByEmail", mock.Anything, "login@example.com").Return(existingUser, nil)

		// Ekspektasi: Redis menyimpan sesi token selama 24 jam
		mockRedis.Regexp().ExpectSet(`^session:.+`, `.+`, 24*time.Hour).SetVal("OK")

		user, token, err := userSvc.SignIn(ctx, "login@example.com", password)

		require.NoError(t, err)
		require.NotEmpty(t, token)
		require.NotNil(t, user)
		assert.Equal(t, int64(10), user.ID)
		assert.Empty(t, user.Password, "Password hash tidak boleh dikembalikan ke client")
	})

	t.Run("gagal login karena password salah", func(t *testing.T) {
		mockRepo, _, _, userSvc, _ := setupServiceTest(t)

		hashed, _ := conv.HashPassword("passwordBenar")
		existingUser := &entity.UserEntity{
			ID:       11,
			Email:    "wrongpass@example.com",
			Password: hashed,
		}

		mockRepo.On("GetUserByEmail", mock.Anything, "wrongpass@example.com").Return(existingUser, nil)

		user, token, err := userSvc.SignIn(ctx, "wrongpass@example.com", "passwordSalah")

		assert.Error(t, err)
		assert.Equal(t, service.ErrInvalidPassword, err)
		assert.Empty(t, token)
		assert.Nil(t, user)
	})

	t.Run("gagal login karena email tidak ditemukan", func(t *testing.T) {
		mockRepo, _, _, userSvc, _ := setupServiceTest(t)

		mockRepo.On("GetUserByEmail", mock.Anything, "notfound@example.com").Return(nil, repository.ErrUserNotFound)

		user, token, err := userSvc.SignIn(ctx, "notfound@example.com", "anyPassword")

		assert.Error(t, err)
		assert.Equal(t, service.ErrInvalidPassword, err)
		assert.Empty(t, token)
		assert.Nil(t, user)
	})
}

func TestUserService_SignUp(t *testing.T) {
	ctx := context.Background()

	t.Run("sukses register user baru", func(t *testing.T) {
		mockRepo, _, _, userSvc, _ := setupServiceTest(t)

		req := entity.UserEntity{
			Name:     "Registran Baru",
			Email:    "new@example.com",
			Password: "securePassword",
		}

		mockRepo.On("GetUserByEmail", mock.Anything, "new@example.com").Return(nil, repository.ErrUserNotFound)
		mockRepo.On("CreateUser", mock.Anything, mock.MatchedBy(func(u entity.UserEntity) bool {
			return u.Email == "new@example.com" && u.Name == "Registran Baru" && u.Password != "securePassword" // harus ter-hash
		})).Return(int64(15), nil)

		err := userSvc.SignUp(ctx, req)
		require.NoError(t, err)
	})

	t.Run("gagal register karena email sudah ada", func(t *testing.T) {
		mockRepo, _, _, userSvc, _ := setupServiceTest(t)

		existing := &entity.UserEntity{
			ID:    1,
			Email: "already@example.com",
		}
		mockRepo.On("GetUserByEmail", mock.Anything, "already@example.com").Return(existing, nil)

		err := userSvc.SignUp(ctx, entity.UserEntity{Email: "already@example.com", Password: "pass"})
		assert.Error(t, err)
		assert.Equal(t, service.ErrUserAlreadyExists, err)
	})
}

func TestUserService_UpdateProfile_CacheInvalidation(t *testing.T) {
	ctx := context.Background()

	t.Run("sukses update profil dan hapus cache di Redis", func(t *testing.T) {
		mockRepo, _, mockRedis, userSvc, _ := setupServiceTest(t)

		req := entity.UserEntity{
			ID:      5,
			Name:    "Nama Diupdate",
			Address: "Alamat Baru",
		}

		mockRepo.On("UpdateDataUser", mock.Anything, mock.Anything).Return(nil)

		// Ekspektasi krusial: Cache profil user:profile:5 harus dihapus dari Redis
		mockRedis.ExpectDel("user:profile:5").SetVal(1)

		err := userSvc.UpdateProfile(ctx, req)
		require.NoError(t, err)
	})

	t.Run("gagal update database (tidak menghapus cache)", func(t *testing.T) {
		mockRepo, _, _, userSvc, _ := setupServiceTest(t)

		req := entity.UserEntity{ID: 5, Name: "Test Fail"}
		mockRepo.On("UpdateDataUser", mock.Anything, mock.Anything).Return(errors.New("db error"))

		err := userSvc.UpdateProfile(ctx, req)
		assert.Error(t, err)
	})
}

func TestUserService_DeleteUser_CacheInvalidation(t *testing.T) {
	ctx := context.Background()

	t.Run("sukses delete user dan bersihkan cache di Redis", func(t *testing.T) {
		mockRepo, _, mockRedis, userSvc, _ := setupServiceTest(t)

		mockRepo.On("DeleteUser", mock.Anything, int64(7)).Return(nil)
		mockRedis.ExpectDel("user:profile:7").SetVal(1)

		err := userSvc.DeleteUser(ctx, 7)
		require.NoError(t, err)
	})
}

func TestUserService_GetAllUsers(t *testing.T) {
	ctx := context.Background()

	t.Run("sukses get all users", func(t *testing.T) {
		mockRepo, _, _, userSvc, _ := setupServiceTest(t)

		query := entity.QueryStringUser{Page: 1, Limit: 10}
		usersList := []entity.UserEntity{
			{ID: 1, Name: "User 1"},
			{ID: 2, Name: "User 2"},
		}

		mockRepo.On("GetAllUsers", mock.Anything, query).Return(usersList, int64(2), nil)

		result, total, err := userSvc.GetAllUsers(ctx, query)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, result, 2)
	})
}
