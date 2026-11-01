package conv_test

import (
	"os"
	"testing"

	"belajar-golang-unit-test/utils/conv"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// 5. BEFORE AND AFTER TEST (Package-level Setup & Teardown via TestMain)
// ============================================================================
// TestMain berjalan SATU KALI sebelum dan sesudah seluruh test pada package ini.
// Berguna untuk menginisialisasi resource global (mock pool, env test, logger, dsb).
func TestMain(m *testing.M) {
	// [BEFORE] Setup global sebelum test dijalankan
	println(">>> [TestMain SETUP] Memulai suite pengujian package conv...")

	// Menjalankan seluruh test pada package
	exitCode := m.Run()

	// [AFTER] Teardown global setelah seluruh test selesai
	println(">>> [TestMain TEARDOWN] Suite pengujian package conv selesai.")
	os.Exit(exitCode)
}

// ============================================================================
// 1. MEMBUAT UNIT TEST & 3. ASSERTIONS (Standar vs Testify)
// ============================================================================
// Aturan fungsi test:
// - Nama diawali 'Test' diikuti huruf kapital
// - Menerima parameter t *testing.T
func TestHashPassword_Basic(t *testing.T) {
	password := "rahasia123"

	// 5. BEFORE AND AFTER TEST (Test-level via t.Cleanup)
	// t.Cleanup dieksekusi saat test ini selesai, bahkan jika test gagal (FailNow).
	t.Cleanup(func() {
		t.Log("Cleanup untuk TestHashPassword_Basic selesai dieksekusi.")
	})

	hash, err := conv.HashPassword(password)

	// Pendekatan Testify Require: Langsung hentikan test jika error terjadi (mirip t.Fatal)
	require.NoError(t, err, "HashPassword seharusnya tidak mengembalikan error")
	require.NotEmpty(t, hash, "Hash hasil tidak boleh kosong")

	// Pendekatan Testify Assert: Catat kegagalan tetapi lanjutkan jika gagal (mirip t.Error)
	assert.NotEqual(t, password, hash, "Hash tidak boleh sama persis dengan plain text")

	// Verifikasi verifikasi hash
	isValid := conv.CheckPasswordHash(password, hash)
	assert.True(t, isValid, "CheckPasswordHash harus bernilai true untuk password yang cocok")

	isInvalid := conv.CheckPasswordHash("salah_password", hash)
	assert.False(t, isInvalid, "CheckPasswordHash harus bernilai false untuk password yang salah")
}

// ============================================================================
// 2. MENGGAGALKAN TEST (t.Fail, t.FailNow, t.Error, t.Fatal)
// ============================================================================
func TestDemonstrasiMenggagalkanTest(t *testing.T) {
	// Contoh ilustrasi logika bagaimana t.Error vs t.Fatal bekerja:
	data := "123"
	num, err := conv.StringToInt64(data)

	if err != nil {
		// t.Fatalf mencatat pesan error lalu menghentikan eksekusi test saat itu juga (FailNow).
		// Sangat penting jika langkah berikutnya bergantung pada data yang valid.
		t.Fatalf("Fatal: konversi string gagal: %v", err)
	}

	// Jika kita sampai di sini, 'num' aman untuk dicek
	if num != 123 {
		// t.Errorf mencatat error tetapi TIDAK menghentikan test (tetap lanjut ke baris berikutnya).
		t.Errorf("Error: didapat %d, ekspektasi 123", num)
	}
}

// ============================================================================
// 4. SKIP TEST (Kondisional & Short Mode)
// ============================================================================
func TestHashingBcrypt_SkipHeavy(t *testing.T) {
	// testing.Short() bernilai true jika perintah dijalankan dengan: go test -short
	if testing.Short() {
		// t.Skip melewati pengujian ini dengan alasan yang jelas
		t.Skip("Melewati pengujian komputasi bcrypt ber-cost tinggi karena mode -short aktif")
	}

	// Uji hashing dengan cost tinggi (hanya dijalankan saat full test)
	hash, err := conv.HashPasswordWithCost("passwordSuperAman", 12)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
}

// ============================================================================
// 6. SUB TEST & 7. TABLE TEST (Table-Driven Tests)
// ============================================================================
// Table test adalah idiom standar Go untuk menguji banyak skenario (happy & unhappy path)
// tanpa menduplikasi kode asersi.
func TestStringToInt64_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{
			name:    "angka positif valid",
			input:   "100",
			want:    100,
			wantErr: false,
		},
		{
			name:    "angka nol",
			input:   "0",
			want:    0,
			wantErr: false,
		},
		{
			name:    "angka negatif valid",
			input:   "-50",
			want:    -50,
			wantErr: false,
		},
		{
			name:    "nilai maksimum int64",
			input:   "9223372036854775807",
			want:    9223372036854775807,
			wantErr: false,
		},
		{
			name:    "string alfabet (harus error)",
			input:   "bukan_angka",
			want:    0,
			wantErr: true,
		},
		{
			name:    "string kosong (harus error)",
			input:   "",
			want:    0,
			wantErr: true,
		},
		{
			name:    "overflow melebihi int64 (harus error)",
			input:   "9999999999999999999999999999999",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		// t.Run membuat SUB TEST independen untuk setiap baris tabel.
		// Memungkinkan menjalankan subtest tertentu dengan flag:
		// go test -v -run TestStringToInt64_TableDriven/angka_positif_valid
		t.Run(tt.name, func(t *testing.T) {
			got, err := conv.StringToInt64(tt.input)

			if tt.wantErr {
				assert.Error(t, err, "Diharapkan menghasilkan error")
			} else {
				assert.NoError(t, err, "Tidak diharapkan menghasilkan error")
				assert.Equal(t, tt.want, got, "Nilai hasil konversi harus sesuai")
			}
		})
	}
}

func TestInt64ToString_TableDriven(t *testing.T) {
	tests := []struct {
		name  string
		input int64
		want  string
	}{
		{name: "positif", input: 42, want: "42"},
		{name: "nol", input: 0, want: "0"},
		{name: "negatif", input: -99, want: "-99"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := conv.Int64ToString(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
