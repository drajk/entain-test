.PHONY: build run run-racing run-api proto-racing proto-api proto clean test kill

export PATH := $(PATH):$(shell go env GOPATH)/bin

build:
	cd racing && go build -o racing .
	cd api && go build -o api .

run: run-racing run-api

run-racing:
	cd racing && go run main.go &

run-api:
	cd api && go run main.go &

proto-racing:
	cd racing && go generate ./proto/...

proto-api:
	cd api && go generate ./proto/...

proto: proto-racing proto-api

clean:
	rm -f racing/racing api/api

test:
	cd racing && go test ./... -v

lint:
	cd racing && golangci-lint run ./...
	cd api && golangci-lint run ./...

kill:
	-lsof -ti :9000 | xargs kill -9 2>/dev/null
	-lsof -ti :8000 | xargs kill -9 2>/dev/null
