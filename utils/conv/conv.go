package conv

import (
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword mengenkripsi plain text password menggunakan bcrypt dengan default cost (10).
func HashPassword(password string) (string, error) {
	return HashPasswordWithCost(password, bcrypt.DefaultCost)
}

// HashPasswordWithCost mengenkripsi plain text password dengan cost tertentu (berguna untuk benchmark).
func HashPasswordWithCost(password string, cost int) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPasswordHash membandingkan plain text password dengan hash bcrypt.
func CheckPasswordHash(password, hash string) bool {
	start := time.Now()
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	log.Debug().
		Dur("duration", time.Since(start)).
		Msg("bcrypt compare finished")
	return err == nil
}

// StringToInt64 mengonversi string ke int64 dengan validasi error.
func StringToInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// Int64ToString mengonversi int64 ke representasi string.
func Int64ToString(n int64) string {
	return strconv.FormatInt(n, 10)
}
