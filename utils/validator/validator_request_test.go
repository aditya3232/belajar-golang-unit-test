package validator_test

import (
	"testing"

	"belajar-golang-unit-test/utils/validator"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sampleUserPayload struct {
	Name     string `validate:"required,min=3"`
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=6"`
}

func TestValidator_Validate_TableDriven(t *testing.T) {
	v := validator.NewValidator()
	require.NotNil(t, v, "Validator instance tidak boleh nil")

	tests := []struct {
		name    string
		payload sampleUserPayload
		wantErr bool
	}{
		{
			name: "payload lengkap dan valid",
			payload: sampleUserPayload{
				Name:     "Aditya",
				Email:    "aditya@example.com",
				Password: "passwordRahasia",
			},
			wantErr: false,
		},
		{
			name: "nama kosong (harus error required)",
			payload: sampleUserPayload{
				Name:     "",
				Email:    "aditya@example.com",
				Password: "passwordRahasia",
			},
			wantErr: true,
		},
		{
			name: "nama kurang dari 3 karakter (harus error min)",
			payload: sampleUserPayload{
				Name:     "Ad",
				Email:    "aditya@example.com",
				Password: "passwordRahasia",
			},
			wantErr: true,
		},
		{
			name: "format email tidak valid (harus error email)",
			payload: sampleUserPayload{
				Name:     "Aditya",
				Email:    "bukan-format-email",
				Password: "passwordRahasia",
			},
			wantErr: true,
		},
		{
			name: "password kurang dari 6 karakter (harus error min)",
			payload: sampleUserPayload{
				Name:     "Aditya",
				Email:    "aditya@example.com",
				Password: "123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.payload)

			if tt.wantErr {
				assert.Error(t, err, "Seharusnya mengembalikan validation error")
			} else {
				assert.NoError(t, err, "Tidak boleh ada validation error")
			}
		})
	}
}
