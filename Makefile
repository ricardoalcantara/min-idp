BINARY := min-idp
CMD    := ./cmd/min-idp

.PHONY: build run tidy test coverage

build:
	go build -o $(BINARY) $(CMD)

run:
	go run $(CMD)

tidy:
	go mod tidy

test:
	go test ./internal/...

coverage:
	go test ./internal/... -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out | grep -E '\.service\.go|total'
