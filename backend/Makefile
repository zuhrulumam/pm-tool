.PHONY: help build dev test lint swagger mock docker-up docker-down migrate-up migrate-down migrate-create clean

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

dev: ## Run with hot reload
	air


test: ## Run tests
	go test -v -cover ./...

test-coverage: ## Run tests with coverage report
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

lint: ## Run linter
	golangci-lint run

swagger: ## Generate swagger documentation
	swag init -g cmd/serve.go

mock: ## Generate mocks for testing
	mockgen -source=business/domain/repository.go -destination=mocks/repository_mock.go

docker-up: ## Start Docker containers
	docker-compose up -d

docker-down: ## Stop Docker containers
	docker-compose down

docker-logs: ## View Docker logs
	docker-compose logs -f

migrate-up: ## Run database migrations up
	go run main.go migrate up

migrate-down: ## Run database migrations down
	go run main.go migrate down

migrate-create: ## Create a new migration (use name=migration_name)
	go run main.go migrate create $(name)

build: ## Build the application
	go build -o bin/pm-tool main.go

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf tmp/
	rm -f coverage.out

deps: ## Download dependencies
	go mod download
	go mod tidy

fmt: ## Format code
	go fmt ./...
	goimports -w .

install-tools: ## Install development tools
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golang/mock/mockgen@latest
