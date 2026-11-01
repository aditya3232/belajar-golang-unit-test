package mocks

import (
	"context"

	"belajar-golang-unit-test/internal/core/domain/entity"

	"github.com/stretchr/testify/mock"
)

// UserRepositoryMock adalah implementasi mock dari UserRepositoryInterface menggunakan testify/mock.
type UserRepositoryMock struct {
	mock.Mock
}

func (m *UserRepositoryMock) CreateUser(ctx context.Context, req entity.UserEntity) (int64, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(int64), args.Error(1)
}

func (m *UserRepositoryMock) GetUserByEmail(ctx context.Context, email string) (*entity.UserEntity, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserEntity), args.Error(1)
}

func (m *UserRepositoryMock) GetUserByID(ctx context.Context, id int64) (*entity.UserEntity, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserEntity), args.Error(1)
}

func (m *UserRepositoryMock) UpdateDataUser(ctx context.Context, req entity.UserEntity) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *UserRepositoryMock) DeleteUser(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *UserRepositoryMock) GetAllUsers(ctx context.Context, query entity.QueryStringUser) ([]entity.UserEntity, int64, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]entity.UserEntity), args.Get(1).(int64), args.Error(2)
}
