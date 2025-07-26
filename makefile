install:
	go mod download
	go mod tidy

run:
	go run cmd/main.go