package conv_test

import (
	"testing"

	"belajar-golang-unit-test/utils/conv"
)

// ============================================================================
// 9. BENCHMARK
// ============================================================================
// Aturan fungsi benchmark:
// - Nama diawali 'Benchmark'
// - Menerima parameter b *testing.B
// - Berisi loop dengan b.N iterasi (diatur otomatis oleh Go runtime)
//
// Dijalankan dengan: go test -bench=. -benchmem ./utils/conv/...
func BenchmarkStringToInt64(b *testing.B) {
	str := "123456789"

	// Melaporkan alokasi memori (B/op dan allocs/op)
	b.ReportAllocs()

	// Mereset penghitung waktu sebelum loop eksekusi utama dimulai

	for b.Loop() {
		_, _ = conv.StringToInt64(str)
	}
}

// ============================================================================
// 10. SUB BENCHMARK
// ============================================================================
// Menggunakan b.Run(name, func(b *testing.B)) untuk membandingkan variasi skenario
// performa dalam satu fungsi benchmark.
//
// Dijalankan spesifik: go test -bench=BenchmarkHashPassword_Costs/cost_4
func BenchmarkHashPassword_Costs(b *testing.B) {
	b.Run("cost_4", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = conv.HashPasswordWithCost("password123", 4)
		}
	})

	b.Run("cost_6", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = conv.HashPasswordWithCost("password123", 6)
		}
	})
}

// ============================================================================
// 11. TABLE BENCHMARK
// ============================================================================
// Menerapkan pola table-driven ke benchmark untuk menguji variasi input ukuran data
// secara sistematis.
func BenchmarkStringToInt64_TableBenchmark(b *testing.B) {
	benchmarks := []struct {
		name  string
		input string
	}{
		{name: "short_number_1_digit", input: "7"},
		{name: "medium_number_5_digits", input: "12345"},
		{name: "large_number_19_digits", input: "9223372036854775800"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = conv.StringToInt64(bm.input)
			}
		})
	}
}
