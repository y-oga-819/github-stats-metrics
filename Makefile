.PHONY: help dev prod build-dev build-prod up-dev up-prod down clean logs test monitoring test-integration test-unit test-all lint-backend build-backend run-backend monitoring-build monitoring-down monitoring-logs monitoring-urls logs-prod health analyze-images prune

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development targets
dev: build-dev up-dev ## Build and start development environment

build-dev: ## Build development containers
	docker compose build

up-dev: ## Start development environment
	docker compose up -d

# Production targets
prod: build-prod up-prod ## Build and start production environment

build-prod: ## Build production containers
	docker compose -f docker compose.prod.yml build

up-prod: ## Start production environment
	docker compose -f docker compose.prod.yml up -d

# Monitoring targets
monitoring: ## Start full monitoring stack (app + prometheus + grafana + loki)
	docker compose -f docker compose.monitoring.yml up -d

monitoring-build: ## Build and start monitoring stack
	docker compose -f docker compose.monitoring.yml up --build -d

monitoring-down: ## Stop monitoring stack
	docker compose -f docker compose.monitoring.yml down

monitoring-logs: ## Show monitoring stack logs
	docker compose -f docker compose.monitoring.yml logs -f

# Monitoring URLs
monitoring-urls: ## Show monitoring service URLs
	@echo "Monitoring Services:"
	@echo "  Application:     http://localhost:3000"
	@echo "  Backend API:     http://localhost:8080"
	@echo "  Prometheus:      http://localhost:9090"
	@echo "  Grafana:         http://localhost:3001 (admin/admin123)"
	@echo "  Loki:            http://localhost:3100"
	@echo ""
	@echo "Health Checks:"
	@echo "  Backend Health:  http://localhost:8080/health"
	@echo "  Backend Metrics: http://localhost:8080/metrics"

# General targets
down: ## Stop all containers
	docker compose down
	docker compose -f docker compose.prod.yml down
	docker compose -f docker compose.monitoring.yml down

clean: ## Remove all containers and images
	docker compose down --rmi all --volumes --remove-orphans
	docker compose -f docker compose.prod.yml down --rmi all --volumes --remove-orphans
	docker compose -f docker compose.monitoring.yml down --rmi all --volumes --remove-orphans

logs: ## Show logs from all containers
	docker compose logs -f

logs-prod: ## Show logs from production containers
	docker compose -f docker compose.prod.yml logs -f

# Health checks
health: ## Check health of all containers
	@echo "Development containers:"
	@docker compose ps
	@echo "\nProduction containers:"
	@docker compose -f docker compose.prod.yml ps
	@echo "\nMonitoring containers:"
	@docker compose -f docker compose.monitoring.yml ps

# Testing
test-backend: ## Run backend tests
	cd backend/app && go test ./...

test-frontend: ## Run frontend tests (if tests exist)
	cd frontend && yarn test --watchAll=false

test: test-backend ## Run all tests

# Integration tests
test-integration: ## Run backend integration tests
	cd backend/app && go test -v ./integration_test/...

test-unit: ## Run only unit tests (excluding integration)
	cd backend/app && go test -v $(shell find . -name "*.go" -path "*/domain/*" -o -path "*/infrastructure/*" | grep -v integration_test | xargs dirname | sort -u | sed 's|^\.|./|')

test-all: test-backend test-integration ## Run all backend tests including integration tests

# Code quality
lint-backend: ## Run Go linter
	cd backend/app && go vet ./...
	cd backend/app && go fmt ./...

build-backend: ## Build backend binary
	cd backend/app && go build -o bin/server cmd/main.go

run-backend: ## Run backend locally (requires environment variables)
	cd backend/app && go run cmd/main.go

# Image size analysis
analyze-images: ## Show image sizes
	@echo "Docker images sizes:"
	@docker images | grep github-stats-metrics

# Cleanup
prune: ## Clean up Docker system
	docker system prune -f
	docker volume prune -f
	docker network prune -f