.PHONY: all build run test test-short test-coverage test-html bench clean

all: test build

build:
	@mkdir -p bin
	go build -o bin/app main.go

run:
	go run main.go start

test:
	go test -v ./...

test-short:
	go test -v -short ./...

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

test-html: test-coverage
	go tool cover -html=coverage.out -o coverage.html
	@echo "Laporan coverage berhasil dibuat di coverage.html"

bench:
	go test -bench=. -benchmem ./...

clean:
	rm -rf bin app coverage.out coverage.html
