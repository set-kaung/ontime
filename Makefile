BINARY_NAME := ontime

run:build
	bin/$(BINARY_NAME)
build:
	sqlc generate & go build -v -o bin/$(BINARY_NAME) ./cmd/

cloud:
	GOOS=linux GOARCH=amd64 go build -v -o bin/ontime_cloud ./cmd/
