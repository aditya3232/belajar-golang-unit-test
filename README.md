# Belajar Unit Testing di Golang (Real-World Project Template)

Proyek ini dibuat sebagai materi pembelajaran **Unit Testing & Benchmarking di Golang** dengan arsitektur riil tingkat industri (*Hexagonal / Clean Layered Architecture*), mengadopsi standar struktur folder dan gaya penulisan dari:
`services/user-service`

Proyek ini mengelola entitas data **User** dengan integrasi **PostgreSQL** (GORM) dan **Redis** (pola *Cache-Aside* dan *Session Store*). Seluruh unit test dirancang **100% Hermetis (Mandiri)** sehingga dapat dijalankan tanpa perlu menyalakan database atau server fisik.

---

## 1. Struktur Direktori

```text
belajar-golang-unit-test/
├── cmd/
│   ├── root.go                          # Cobra CLI root & Viper environment loader
│   └── start.go                         # Subcommand 'start' untuk menjalankan server
├── config/
│   ├── config.go                        # Mapping konfigurasi environment (.env)
│   ├── database.go                      # Inisialisasi pool koneksi GORM PostgreSQL
│   ├── fiber.go                         # Konfigurasi Fiber v3 & custom error handler
│   ├── redis.go                         # Inisialisasi client Redis (go-redis v9)
│   └── zerolog.go                       # Konfigurasi logger Zerolog terstruktur
├── database/
│   └── migrations/
│       ├── 000001_create_users_table.up.sql
│       └── 000001_create_users_table.down.sql
├── internal/
│   ├── adapter/
│   │   ├── handler/
│   │   │   ├── request/user_request.go  # DTO input HTTP & tag validator
│   │   │   ├── response/                # DTO output HTTP & pagination format
│   │   │   ├── user_handler.go          # HTTP Handlers (SignUp, SignIn, Profile, dsb.)
│   │   │   └── user_handler_test.go     # [TEST] Pengujian Handler via Fiber app.Test()
│   │   ├── middleware_adapter.go        # Middleware autentikasi Bearer JWT & Redis Session
│   │   ├── middleware_adapter_test.go   # [TEST] Pengujian Middleware (401 vs 200)
│   │   └── repository/
│   │       ├── user_repository.go       # Implementasi GORM CRUD database
│   │       └── user_repository_test.go  # [TEST] Pengujian Query DB via DATA-DOG/go-sqlmock
│   ├── app/
│   │   └── app.go                       # Dependency Injection, routes wiring & graceful shutdown
│   └── core/
│       ├── domain/
│       │   ├── entity/user_entity.go    # Pure business domain entities
│       │   ├── entity/jwt_entity.go     # DTO JWT User Session
│       │   └── model/user_model.go      # Schema GORM Model & Table Mapping
│       └── service/
│           ├── jwt_service.go           # Penandatanganan & verifikasi token JWT
│           ├── jwt_service_test.go      # [TEST] Unit test JWT signing & claims
│           ├── user_service.go          # Logika Bisnis: Cache-Aside Redis & DB
│           ├── user_service_test.go     # [TEST] Unit test Service via Testify Mock & Redismock
│           └── mocks/
│               └── user_repository_mock.go # Mock UserRepository berbasis testify/mock
├── utils/
│   ├── conv/
│   │   ├── conv.go                      # Helper Bcrypt hashing & konversi int/string
│   │   ├── conv_test.go                 # [TEST] Demonstrasi 7 konsep fundamental test
│   │   └── conv_benchmark_test.go       # [BENCHMARK] Demonstrasi 3 konsep benchmark
│   └── validator/
│       ├── validator_request.go         # Wrapper validator v10 & translator
│       └── validator_request_test.go    # [TEST] Table-driven test validasi input
├── .env.example                         # Contoh variabel lingkungan
├── go.mod                               # Daftar modul & dependensi Go
├── issue.md                             # Dokumen perencanaan & materi pengujian
├── main.go                              # Entry point aplikasi
└── Makefile                             # Shortcut perintah test, bench, cover, build
```

---

## 2. 11 Fundamental Unit Testing & Benchmark di Go

Materi fundamental ini diimplementasikan secara langsung di dalam kode:

| No | Konsep Fundamental | Lokasi Contoh Kode | Penjelasan Singkat |
| :---: | :--- | :--- | :--- |
| 1 | **Membuat Unit Test** | `utils/conv/conv_test.go` | Penamaan `*_test.go`, signature `func TestXxx(t *testing.T)`. |
| 2 | **Menggagalkan Test** | `utils/conv/conv_test.go` | Perbedaan `t.Fail()` vs `t.FailNow()` dan `t.Error()` vs `t.Fatal()`. |
| 3 | **Assertions** | `utils/conv/conv_test.go` | Idiom `if !=` standar Go vs `assert.Equal` & `require.NoError` (`testify`). |
| 4 | **Skip Test** | `utils/conv/conv_test.go` | Melewatkan test (`t.Skip`) dan penggunaan flag `testing.Short()`. |
| 5 | **Before & After Test** | `utils/conv/conv_test.go` | Package-level setup/teardown (`TestMain`) & test-level cleanup (`t.Cleanup`). |
| 6 | **Sub Test** | `utils/conv/conv_test.go` | Pengelompokan pengujian hierarkis dengan `t.Run(name, func)`. |
| 7 | **Table Test** | `utils/conv/conv_test.go` | Idiom table-driven tests (`tests := []struct{...}`) untuk beragam kombinasi. |
| 8 | **Mocking** | `internal/core/service/user_service_test.go` | Isolasi repository & Redis via `testify/mock` dan `go-redis/redismock/v9`. |
| 9 | **Benchmark** | `utils/conv/conv_benchmark_test.go` | Pengukuran kecepatan operasi (`b *testing.B`, `b.N`, `b.ReportAllocs()`). |
| 10 | **Sub Benchmark** | `utils/conv/conv_benchmark_test.go` | Membandingkan performa variasi input dengan `b.Run()`. |
| 11 | **Table Benchmark** | `utils/conv/conv_benchmark_test.go` | Pengujian performa sistematis berbasis tabel input ukuran data. |

---

## 3. Strategi Pengujian per Lapisan

### 3.1. Repository Layer (`go-sqlmock`)
- Menguji query database GORM (`SELECT`, `INSERT`, `UPDATE`, `DELETE`) dan transaksi (`BEGIN`, `COMMIT`, `ROLLBACK`) tanpa database PostgreSQL fisik.
- Berkas: `internal/adapter/repository/user_repository_test.go`

### 3.2. Service Layer (`testify/mock` & `redismock/v9`)
- Menguji logika bisnis **Cache-Aside Redis**:
  - **Cache Hit**: Data diambil langsung dari Redis. Verifikasi bahwa database repository **TIDAK** dipanggil.
  - **Cache Miss**: Redis return nil -> Database query dipanggil -> Hasil disimpan ke Redis dengan TTL 15 menit.
  - **Cache Invalidation**: Saat Update atau Delete, cache `user:profile:<id>` dihapus (`DEL`) dari Redis.
- Berkas: `internal/core/service/user_service_test.go`

### 3.3. Middleware Layer (`redismock/v9` + JWT)
- Menguji verifikasi Header `Authorization: Bearer <token>`, validasi tanda tangan JWT, dan pengecekan sesi aktif di Redis (`session:<token>`).
- Berkas: `internal/adapter/middleware_adapter_test.go`

### 3.4. Handler Layer (`Fiber app.Test()`)
- Menguji HTTP status code (200, 201, 400, 401, 404, 409), deserialisasi JSON response `response.DefaultResponse`, dan validasi payload tanpa membuka port jaringan.
- Berkas: `internal/adapter/handler/user_handler_test.go`

---

## 4. Cara Menjalankan Pengujian & Benchmark

Gunakan `make` untuk menjalankan perintah-perintah praktis berikut:

### 1. Menjalankan Seluruh Unit Test
```bash
make test
# atau
go test -v ./...
```

### 2. Menjalankan Test Cepat (Mode Short)
```bash
make test-short
# atau
go test -v -short ./...
```

### 3. Menghitung Code Coverage
```bash
make test-coverage
# atau
go test -v -coverprofile=coverage.out ./... && go tool cover -func=coverage.out
```

### 4. Menghasilkan Laporan Visual HTML Coverage
```bash
make test-html
```
Perintah ini akan membuat berkas `coverage.html` yang dapat dibuka di browser untuk melihat baris kode mana saja yang ter-cover oleh unit test.

### 5. Menjalankan Seluruh Benchmark & Memory Profiling
```bash
make bench
# atau
go test -bench=. -benchmem ./...
```
Output akan menampilkan:
- `ns/op`: Waktu yang dibutuhkan per operasi dalam nanodetik.
- `B/op`: Jumlah alokasi memori heap per operasi dalam bytes.
- `allocs/op`: Frekuensi alokasi memori heap per operasi.
