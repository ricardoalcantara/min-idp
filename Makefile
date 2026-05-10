BINARY := min-idp
CMD    := ./cmd/min-idp

.PHONY: build run tidy

build:
	go build -o $(BINARY) $(CMD)

run:
	go run $(CMD)

tidy:
	go mod tidy
