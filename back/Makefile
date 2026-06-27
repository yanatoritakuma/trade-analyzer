# back/Makefile

# openapi.yaml → Goコード自動生成
generate-go:
	oapi-codegen -generate types -o internal/api/types.gen.go -package api openapi.yaml
	oapi-codegen -generate gin-server -o internal/api/server.gen.go -package api openapi.yaml

# DBマイグレーション実行
migrate:
	go run migrate/migrate.go

# Docker通常起動（ローカルDB込み）
up:
	docker compose up --build

# Dockerホットリロード起動（開発・ローカルDB込み）
dev:
	docker compose -f docker-compose.dev.yml up --build

# Docker停止
down:
	docker compose down

# テスト
test:
	go test ./... -v

test-unit:
	go test ./domain/... ./usecase/... -v

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

# AWS Lambda向けGoバイナリビルド
build-lambda:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap -ldflags="-s -w" .
	rm -f lambda-handler.zip
	zip lambda-handler.zip bootstrap
	rm bootstrap
