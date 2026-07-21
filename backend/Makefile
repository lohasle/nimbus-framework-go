.PHONY: tidy swagger build test run frontend-build

GOPROXY ?= https://goproxy.cn,direct

tidy:
	GOPROXY=$(GOPROXY) go mod tidy

swagger:
	GOPROXY=$(GOPROXY) go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/server/main.go -o docs --parseDependency

build: swagger
	mkdir -p bin
	go build -trimpath -o bin/nimbus-server ./cmd/server

test:
	go test ./...

run:
	go run ./cmd/server

frontend-build:
	cd ../frontend && corepack pnpm install --frozen-lockfile && corepack pnpm build:prod
