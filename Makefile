BINARY_NAME := ontime

run:build
	bin/$(BINARY_NAME)
build:
	go build -o bin/$(BINARY_NAME) ./cmd/

