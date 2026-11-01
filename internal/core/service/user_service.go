package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"belajar-golang-unit-test/config"
	"belajar-golang-unit-test/internal/adapter/repository"
	"belajar-golang-unit-test/internal/core/domain/entity"
	"belajar-golang-unit-test/utils/conv"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

var (
	ErrUserAlreadyExists = errors.New("email already registered")
	ErrInvalidPassword   = errors.New("invalid email or password")
)

type UserServiceInterface interface {
	SignIn(ctx context.Context, email, password string) (*entity.UserEntity, string, error)
	SignUp(ctx context.Context, req entity.UserEntity) error
	GetProfile(ctx context.Context, userID int64) (*entity.UserEntity, error)
	UpdateProfile(ctx context.Context, req entity.UserEntity) error
	DeleteUser(ctx context.Context, userID int64) error
	GetAllUsers(ctx context.Context, query entity.QueryStringUser) ([]entity.UserEntity, int64, error)
}

type userService struct {
	repo       repository.UserRepositoryInterface
	jwtService JwtServiceInterface
	redis      *redis.Client
	cfg        *config.Config
}

func NewUserService(
	repo repository.UserRepositoryInterface,
	jwtService JwtServiceInterface,
	redisClient *redis.Client,
	cfg *config.Config,
) UserServiceInterface {
	return &userService{
		repo:       repo,
		jwtService: jwtService,
		redis:      redisClient,
		cfg:        cfg,
	}
}

func (s *userService) SignIn(ctx context.Context, email, password string) (*entity.UserEntity, string, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, "", ErrInvalidPassword
		}
		return nil, "", err
	}

	if !conv.CheckPasswordHash(password, user.Password) {
		return nil, "", ErrInvalidPassword
	}

	token, err := s.jwtService.GenerateToken(user.ID, user.Email, user.Name, user.Role)
	if err != nil {
		return nil, "", err
	}

	sessionData := entity.JwtUserData{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
		Role:  user.Role,
	}

	jsonData, err := json.Marshal(sessionData)
	if err == nil && s.redis != nil {
		sessionKey := fmt.Sprintf("session:%s", token)
		_ = s.redis.Set(ctx, sessionKey, jsonData, 24*time.Hour).Err()
	}

	user.Password = ""
	return user, token, nil
}

func (s *userService) SignUp(ctx context.Context, req entity.UserEntity) error {
	existing, _ := s.repo.GetUserByEmail(ctx, req.Email)
	if existing != nil {
		return ErrUserAlreadyExists
	}

	hashed, err := conv.HashPassword(req.Password)
	if err != nil {
		return err
	}

	req.Password = hashed
	if req.Role == "" {
		req.Role = "user"
	}

	_, err = s.repo.CreateUser(ctx, req)
	return err
}

func (s *userService) GetProfile(ctx context.Context, userID int64) (*entity.UserEntity, error) {
	cacheKey := fmt.Sprintf("user:profile:%d", userID)

	// 1. Periksa cache Redis terlebih dahulu (Cache Hit)
	if s.redis != nil {
		cachedStr, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil && cachedStr != "" {
			var cachedUser entity.UserEntity
			if err := json.Unmarshal([]byte(cachedStr), &cachedUser); err == nil {
				log.Debug().Int64("user_id", userID).Msg("cache hit for user profile")
				return &cachedUser, nil
			}
		}
	}

	// 2. Cache Miss: Ambil dari database PostgreSQL
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 3. Simpan data ke Redis dengan TTL 15 menit
	if s.redis != nil {
		userBytes, err := json.Marshal(user)
		if err == nil {
			_ = s.redis.Set(ctx, cacheKey, userBytes, 15*time.Minute).Err()
		}
	}

	return user, nil
}

func (s *userService) UpdateProfile(ctx context.Context, req entity.UserEntity) error {
	if req.Password != "" {
		hashed, err := conv.HashPassword(req.Password)
		if err != nil {
			return err
		}
		req.Password = hashed
	}

	if err := s.repo.UpdateDataUser(ctx, req); err != nil {
		return err
	}

	// Invalidate Cache di Redis
	if s.redis != nil {
		cacheKey := fmt.Sprintf("user:profile:%d", req.ID)
		_ = s.redis.Del(ctx, cacheKey).Err()
	}

	return nil
}

func (s *userService) DeleteUser(ctx context.Context, userID int64) error {
	if err := s.repo.DeleteUser(ctx, userID); err != nil {
		return err
	}

	// Invalidate Cache di Redis
	if s.redis != nil {
		cacheKey := fmt.Sprintf("user:profile:%d", userID)
		_ = s.redis.Del(ctx, cacheKey).Err()
	}

	return nil
}

func (s *userService) GetAllUsers(ctx context.Context, query entity.QueryStringUser) ([]entity.UserEntity, int64, error) {
	return s.repo.GetAllUsers(ctx, query)
}
