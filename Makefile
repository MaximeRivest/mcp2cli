.PHONY: fmt test build tidy smoke

fmt:
	gofmt -w ./cmd ./internal

test:
	go test ./...

build:
	go build -o bin/mcptocli ./cmd/mcptocli

tidy:
	go mod tidy

smoke:
	bash scripts/smoke.sh
