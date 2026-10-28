.PHONY: up down migrate-all seed-dev test lint build run-identity run-catalog run-order run-payment run-notification images push

up:
	docker compose up -d

down:
	docker compose down

build:
	@for svc in identity catalog order payment notification; do \
		echo "Building $$svc..."; \
		cd services/$$svc && go build -o /dev/null ./cmd/server && cd ../..; \
	done
	@echo "All services build successfully"

test:
	@for svc in identity catalog order payment notification; do \
		echo "Testing $$svc..."; \
		cd services/$$svc && go test ./... && cd ../..; \
	done
	cd pkg && go test ./...

lint:
	@for svc in identity catalog order payment notification; do \
		echo "Linting $$svc..."; \
		cd services/$$svc && golangci-lint run ./... && cd ../..; \
	done

migrate-all:
	@echo "Running Identity migrations..."
	migrate -path services/identity/migrations -database "$$IDENTITY_DB_URL" up
	@echo "Running Catalog migrations..."
	migrate -path services/catalog/migrations -database "$$CATALOG_DB_URL" up
	@echo "Running Order migrations..."
	migrate -path services/order/migrations -database "$$ORDER_DB_URL" up
	@echo "Running Payment migrations..."
	migrate -path services/payment/migrations -database "$$PAYMENT_DB_URL" up

seed-dev:
	@echo "Seeding dev data..."
	psql "$$IDENTITY_DB_URL" -v ON_ERROR_STOP=1 -f infra/seed/identity.sql
	psql "$$CATALOG_DB_URL" -v ON_ERROR_STOP=1 -f infra/seed/catalog.sql
	psql "$$PAYMENT_DB_URL" -v ON_ERROR_STOP=1 -f infra/seed/payment.sql

run-identity:
	cd services/identity && go run ./cmd/server

run-catalog:
	cd services/catalog && go run ./cmd/server

run-order:
	cd services/order && go run ./cmd/server

run-payment:
	cd services/payment && go run ./cmd/server

run-notification:
	cd services/notification && go run ./cmd/server

# Build and push prod images: make push IMAGE_PREFIX=docker.io/yourname IMAGE_TAG=v1
IMAGE_PREFIX ?= docker.io/CHANGE_ME
IMAGE_TAG ?= latest

images:
	@for svc in identity catalog order payment notification; do \
		echo "Building $(IMAGE_PREFIX)/quickeats-$$svc:$(IMAGE_TAG)"; \
		docker build -f services/$$svc/Dockerfile -t $(IMAGE_PREFIX)/quickeats-$$svc:$(IMAGE_TAG) . || exit 1; \
	done

push: images
	@for svc in identity catalog order payment notification; do \
		docker push $(IMAGE_PREFIX)/quickeats-$$svc:$(IMAGE_TAG) || exit 1; \
	done
