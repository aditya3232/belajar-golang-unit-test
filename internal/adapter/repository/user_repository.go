package repository

import (
	"context"
	"errors"
	"time"

	"belajar-golang-unit-test/internal/core/domain/entity"
	"belajar-golang-unit-test/internal/core/domain/model"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound = errors.New("404")
)

type userRepository struct {
	db *gorm.DB
}

type UserRepositoryInterface interface {
	CreateUser(ctx context.Context, req entity.UserEntity) (int64, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.UserEntity, error)
	GetUserByID(ctx context.Context, id int64) (*entity.UserEntity, error)
	UpdateDataUser(ctx context.Context, req entity.UserEntity) error
	DeleteUser(ctx context.Context, id int64) error
	GetAllUsers(ctx context.Context, query entity.QueryStringUser) ([]entity.UserEntity, int64, error)
}

func NewUserRepository(db *gorm.DB) UserRepositoryInterface {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, req entity.UserEntity) (int64, error) {
	now := time.Now()
	userModel := model.User{
		Name:      req.Name,
		Email:     req.Email,
		Password:  req.Password,
		Phone:     req.Phone,
		Address:   req.Address,
		Role:      req.Role,
		CreatedAt: now,
		UpdatedAt: now,
	}

	result := r.db.WithContext(ctx).Create(&userModel)
	if result.Error != nil {
		log.Error().
			Err(result.Error).
			Str("source", "repository.CreateUser").
			Msg("failed to insert user")
		return 0, result.Error
	}

	return userModel.ID, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*entity.UserEntity, error) {
	var userModel model.User

	result := r.db.WithContext(ctx).Where("email = ?", email).First(&userModel)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		log.Error().
			Err(result.Error).
			Str("source", "repository.GetUserByEmail").
			Msg("failed to get user by email")
		return nil, result.Error
	}

	return &entity.UserEntity{
		ID:        userModel.ID,
		Name:      userModel.Name,
		Email:     userModel.Email,
		Password:  userModel.Password,
		Phone:     userModel.Phone,
		Address:   userModel.Address,
		Role:      userModel.Role,
		CreatedAt: userModel.CreatedAt.Format(time.RFC3339),
		UpdatedAt: userModel.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id int64) (*entity.UserEntity, error) {
	var userModel model.User

	result := r.db.WithContext(ctx).Where("id = ?", id).First(&userModel)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		log.Error().
			Err(result.Error).
			Str("source", "repository.GetUserByID").
			Msg("failed to get user by id")
		return nil, result.Error
	}

	return &entity.UserEntity{
		ID:        userModel.ID,
		Name:      userModel.Name,
		Email:     userModel.Email,
		Password:  userModel.Password,
		Phone:     userModel.Phone,
		Address:   userModel.Address,
		Role:      userModel.Role,
		CreatedAt: userModel.CreatedAt.Format(time.RFC3339),
		UpdatedAt: userModel.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (r *userRepository) UpdateDataUser(ctx context.Context, req entity.UserEntity) error {
	updates := map[string]any{
		"name":       req.Name,
		"phone":      req.Phone,
		"address":    req.Address,
		"updated_at": time.Now(),
	}

	if req.Password != "" {
		updates["password"] = req.Password
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}

	result := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", req.ID).Updates(updates)
	if result.Error != nil {
		log.Error().
			Err(result.Error).
			Str("source", "repository.UpdateDataUser").
			Msg("failed to update user")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *userRepository) DeleteUser(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.User{})
	if result.Error != nil {
		log.Error().
			Err(result.Error).
			Str("source", "repository.DeleteUser").
			Msg("failed to delete user")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *userRepository) GetAllUsers(ctx context.Context, query entity.QueryStringUser) ([]entity.UserEntity, int64, error) {
	var users []model.User
	var total int64

	dbQuery := r.db.WithContext(ctx).Model(&model.User{})

	if query.Search != "" {
		search := "%" + query.Search + "%"
		dbQuery = dbQuery.Where("name ILIKE ? OR email ILIKE ?", search, search)
	}

	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := query.Limit
	if limit <= 0 {
		limit = 10
	}
	page := query.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	order := "id DESC"
	if query.OrderBy != "" {
		orderType := "ASC"
		if query.OrderType == "DESC" || query.OrderType == "desc" {
			orderType = "DESC"
		}
		order = query.OrderBy + " " + orderType
	}

	if err := dbQuery.Order(order).Limit(int(limit)).Offset(int(offset)).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	result := make([]entity.UserEntity, 0, len(users))
	for _, u := range users {
		result = append(result, entity.UserEntity{
			ID:        u.ID,
			Name:      u.Name,
			Email:     u.Email,
			Phone:     u.Phone,
			Address:   u.Address,
			Role:      u.Role,
			CreatedAt: u.CreatedAt.Format(time.RFC3339),
			UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
		})
	}

	return result, total, nil
}
