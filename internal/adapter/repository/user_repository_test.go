package repository_test

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"belajar-golang-unit-test/internal/adapter/repository"
	"belajar-golang-unit-test/internal/core/domain/entity"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupMockDB membuat instance mock database PostgreSQL in-memory menggunakan sqlmock
func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "sqlmock.New() tidak boleh error")

	dialector := postgres.New(postgres.Config{
		Conn:       db,
		DriverName: "postgres",
	})

	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err, "gorm.Open() tidak boleh error")

	// Pastikan semua ekspektasi query sqlmock terpenuhi di akhir pengujian
	t.Cleanup(func() {
		assert.NoError(t, mock.ExpectationsWereMet(), "Semua ekspektasi query sqlmock harus terpenuhi")
	})

	return gormDB, mock
}

func TestUserRepository_GetUserByID(t *testing.T) {
	ctx := context.Background()

	t.Run("sukses mendapatkan data user", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := repository.NewUserRepository(gormDB)

		rows := sqlmock.NewRows([]string{"id", "name", "email", "password", "phone", "address", "role", "created_at", "updated_at"}).
			AddRow(1, "Adit", "adit@example.com", "hashPassword", "08123456789", "Jakarta", "user", time.Now(), time.Now())

		// Query yang diharapkan oleh GORM
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE id = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
			WithArgs(int64(1), 1).
			WillReturnRows(rows)

		user, err := repo.GetUserByID(ctx, 1)

		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, int64(1), user.ID)
		assert.Equal(t, "Adit", user.Name)
		assert.Equal(t, "adit@example.com", user.Email)
	})

	t.Run("user tidak ditemukan (record not found)", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := repository.NewUserRepository(gormDB)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE id = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
			WithArgs(int64(999), 1).
			WillReturnError(gorm.ErrRecordNotFound)

		user, err := repo.GetUserByID(ctx, 999)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, repository.ErrUserNotFound, err)
	})

	t.Run("database error koneksi", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := repository.NewUserRepository(gormDB)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE id = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
			WithArgs(int64(1), 1).
			WillReturnError(errors.New("db connection timeout"))

		user, err := repo.GetUserByID(ctx, 1)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "connection timeout")
	})
}

func TestUserRepository_GetUserByEmail(t *testing.T) {
	ctx := context.Background()

	t.Run("sukses mendapatkan user by email", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := repository.NewUserRepository(gormDB)

		rows := sqlmock.NewRows([]string{"id", "name", "email", "password", "phone", "address", "role", "created_at", "updated_at"}).
			AddRow(2, "Budi", "budi@example.com", "hashBudi", "08987654321", "Bandung", "user", time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
			WithArgs("budi@example.com", 1).
			WillReturnRows(rows)

		user, err := repo.GetUserByEmail(ctx, "budi@example.com")

		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, int64(2), user.ID)
		assert.Equal(t, "budi@example.com", user.Email)
	})

	t.Run("email tidak terdaftar", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := repository.NewUserRepository(gormDB)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
			WithArgs("unknown@example.com", 1).
			WillReturnError(gorm.ErrRecordNotFound)

		user, err := repo.GetUserByEmail(ctx, "unknown@example.com")

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, repository.ErrUserNotFound, err)
	})
}

func TestUserRepository_CreateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("sukses insert data user", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := repository.NewUserRepository(gormDB)

		req := entity.UserEntity{
			Name:     "Citra",
			Email:    "citra@example.com",
			Password: "hashPasswordCitra",
			Phone:    "0811223344",
			Address:  "Surabaya",
			Role:     "user",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
		mock.ExpectCommit()

		id, err := repo.CreateUser(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, int64(10), id)
	})

	t.Run("gagal insert (duplicate key error)", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := repository.NewUserRepository(gormDB)

		req := entity.UserEntity{
			Name:     "Citra",
			Email:    "citra@example.com",
			Password: "hashPasswordCitra",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).
			WillReturnError(errors.New("duplicate key value violates unique constraint"))
		mock.ExpectRollback()

		id, err := repo.CreateUser(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, int64(0), id)
	})
}

func TestUserRepository_UpdateDataUser(t *testing.T) {
	ctx := context.Background()

	t.Run("sukses update profil", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := repository.NewUserRepository(gormDB)

		req := entity.UserEntity{
			ID:      1,
			Name:    "Adit Update",
			Phone:   "08999999",
			Address: "Jakarta Barat",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users" SET`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdateDataUser(ctx, req)
		require.NoError(t, err)
	})

	t.Run("user yang diupdate tidak ada (0 rows affected)", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := repository.NewUserRepository(gormDB)

		req := entity.UserEntity{
			ID:      999,
			Name:    "User Ghaib",
			Phone:   "08000000",
			Address: "Nowhere",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users" SET`)).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		err := repo.UpdateDataUser(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, repository.ErrUserNotFound, err)
	})
}

func TestUserRepository_DeleteUser(t *testing.T) {
	ctx := context.Background()

	t.Run("sukses soft delete user", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := repository.NewUserRepository(gormDB)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users" SET "deleted_at"=$1 WHERE id = $2 AND "users"."deleted_at" IS NULL`)).
			WithArgs(sqlmock.AnyArg(), int64(1)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.DeleteUser(ctx, 1)
		require.NoError(t, err)
	})

	t.Run("user tidak ditemukan saat dihapus", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := repository.NewUserRepository(gormDB)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users" SET "deleted_at"=$1 WHERE id = $2 AND "users"."deleted_at" IS NULL`)).
			WithArgs(sqlmock.AnyArg(), int64(999)).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		err := repo.DeleteUser(ctx, 999)
		assert.Error(t, err)
		assert.Equal(t, repository.ErrUserNotFound, err)
	})
}

func TestUserRepository_GetAllUsers(t *testing.T) {
	ctx := context.Background()

	t.Run("sukses mengambil list user dengan count", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := repository.NewUserRepository(gormDB)

		// Mock Count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "users" WHERE "users"."deleted_at" IS NULL`)).
			WillReturnRows(countRows)

		// Mock Find query
		dataRows := sqlmock.NewRows([]string{"id", "name", "email", "password", "phone", "address", "role", "created_at", "updated_at"}).
			AddRow(1, "User Satu", "satu@example.com", "hash", "081", "Alamat 1", "user", time.Now(), time.Now()).
			AddRow(2, "User Dua", "dua@example.com", "hash", "082", "Alamat 2", "user", time.Now(), time.Now())
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."deleted_at" IS NULL ORDER BY id DESC LIMIT $1`)).
			WithArgs(10).
			WillReturnRows(dataRows)

		users, total, err := repo.GetAllUsers(ctx, entity.QueryStringUser{})
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, users, 2)
	})
}
