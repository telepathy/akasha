.PHONY: run build test clean deps tidy vendor

run:
	go run -mod=vendor cmd/server/main.go

build:
	go build -mod=vendor -o bin/akasha cmd/server/main.go

test:
	go test -mod=vendor ./...

clean:
	rm -rf bin/

deps:
	go mod download

vendor:
	go mod vendor

tidy:
	go mod tidy

lint:
	golangci-lint run