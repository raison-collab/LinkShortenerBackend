.PHONY: all build run test clean docker-build docker-run docker-stop migrate swagger

# Variables
BINARY_NAME=link-shortener
DOCKER_IMAGE=link-shortener:latest
GO=go
GOFLAGS=-v

build:
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) cmd/api/main.go

run:
	$(GO) run cmd/api/main.go

test:
	$(GO) test $(GOFLAGS) -race -coverprofile=coverage.out ./...

test-coverage: test
	$(GO) tool cover -html=coverage.out

clean:
	$(GO) clean
	rm -f $(BINARY_NAME)
	rm -f coverage.out

deps:
	$(GO) mod download
	$(GO) mod tidy

# Run linter
lint:
	golangci-lint run

# Generate swagger documentation
swagger:
	swag init -g cmd/api/main.go

# Docker commands
docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down

docker-logs:
	docker-compose logs -f app

# Database migrations (TODO: подумать над тем, чтобы подсасывать переменные из .env)
migrate-up:
	docker-compose exec postgres psql -U postgres -d link_shortener -f /docker-entrypoint-initdb.d/001_create_users_table.sql
	docker-compose exec postgres psql -U postgres -d link_shortener -f /docker-entrypoint-initdb.d/002_create_links_table.sql
	docker-compose exec postgres psql -U postgres -d link_shortener -f /docker-entrypoint-initdb.d/003_create_link_clicks_table.sql

# Development setup
dev-setup: deps docker-run migrate-up
	@echo "Development environment is ready!"
	@echo "API is running at http://localhost:8080"
	@echo "Swagger docs at http://localhost:8080/swagger/index.html"


fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

# All checks
check: fmt vet lint test

# Default target
all: clean deps build