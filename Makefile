APP_NAME := bitely-api
PORT ?= 8080

.PHONY: run docker-build docker-run docker-up docker-down migrate-up migrate-down

run:
	go run ./cmd/server

docker-build:
	docker build -t $(APP_NAME) .

docker-run:
	@if [ -z "$(DATABASE_URL)" ]; then echo "DATABASE_URL is required"; exit 1; fi
	@if [ -z "$(JWT_SECRET)" ]; then echo "JWT_SECRET is required"; exit 1; fi
	@if [ -z "$(R2_PUBLIC_BASE_URL)" ]; then echo "R2_PUBLIC_BASE_URL is required"; exit 1; fi
	@if [ -z "$(R2_ACCOUNT_ID)" ]; then echo "R2_ACCOUNT_ID is required"; exit 1; fi
	@if [ -z "$(R2_ACCESS_KEY_ID)" ]; then echo "R2_ACCESS_KEY_ID is required"; exit 1; fi
	@if [ -z "$(R2_SECRET_ACCESS_KEY)" ]; then echo "R2_SECRET_ACCESS_KEY is required"; exit 1; fi
	@if [ -z "$(R2_BUCKET)" ]; then echo "R2_BUCKET is required"; exit 1; fi
	docker run --name $(APP_NAME) --rm \
		-p $(PORT):8080 \
		-e PORT=8080 \
		-e DATABASE_URL="$(DATABASE_URL)" \
		-e JWT_SECRET="$(JWT_SECRET)" \
		-e R2_PUBLIC_BASE_URL="$(R2_PUBLIC_BASE_URL)" \
		-e R2_ACCOUNT_ID="$(R2_ACCOUNT_ID)" \
		-e R2_ACCESS_KEY_ID="$(R2_ACCESS_KEY_ID)" \
		-e R2_SECRET_ACCESS_KEY="$(R2_SECRET_ACCESS_KEY)" \
		-e R2_BUCKET="$(R2_BUCKET)" \
		$(APP_NAME)

docker-up: docker-build docker-run

docker-down:
	-docker stop $(APP_NAME)

migrate-up:
	@if [ -z "$(DATABASE_URL)" ]; then echo "DATABASE_URL is required"; exit 1; fi
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	@if [ -z "$(DATABASE_URL)" ]; then echo "DATABASE_URL is required"; exit 1; fi
	migrate -path migrations -database "$(DATABASE_URL)" down 1