.PHONY: help dev prod build-dev build-prod up-dev up-prod down clean logs test monitoring

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development targets
dev: build-dev up-dev ## Build and start development environment

build-dev: ## Build development containers
	docker-compose build

up-dev: ## Start development environment
	docker-compose up -d

# Production targets
prod: build-prod up-prod ## Build and start production environment

build-prod: ## Build production containers
	docker-compose -f docker-compose.prod.yml build

up-prod: ## Start production environment
	docker-compose -f docker-compose.prod.yml up -d

# Monitoring targets
monitoring: ## Start full monitoring stack (app + prometheus + grafana + loki)
	docker-compose -f docker-compose.monitoring.yml up -d

monitoring-build: ## Build and start monitoring stack
	docker-compose -f docker-compose.monitoring.yml up --build -d

monitoring-down: ## Stop monitoring stack
	docker-compose -f docker-compose.monitoring.yml down

monitoring-logs: ## Show monitoring stack logs
	docker-compose -f docker-compose.monitoring.yml logs -f

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
	docker-compose down
	docker-compose -f docker-compose.prod.yml down
	docker-compose -f docker-compose.monitoring.yml down

clean: ## Remove all containers and images
	docker-compose down --rmi all --volumes --remove-orphans
	docker-compose -f docker-compose.prod.yml down --rmi all --volumes --remove-orphans
	docker-compose -f docker-compose.monitoring.yml down --rmi all --volumes --remove-orphans

logs: ## Show logs from all containers
	docker-compose logs -f

logs-prod: ## Show logs from production containers
	docker-compose -f docker-compose.prod.yml logs -f

# Health checks
health: ## Check health of all containers
	@echo "Development containers:"
	@docker-compose ps
	@echo "\nProduction containers:"
	@docker-compose -f docker-compose.prod.yml ps
	@echo "\nMonitoring containers:"
	@docker-compose -f docker-compose.monitoring.yml ps

# Testing
test-backend: ## Run backend tests
	cd backend/app && go test ./...

test-frontend: ## Run frontend tests (if tests exist)
	cd frontend && yarn test --watchAll=false

test: test-backend ## Run all tests

# Image size analysis
analyze-images: ## Show image sizes
	@echo "Docker images sizes:"
	@docker images | grep github-stats-metrics

# Cleanup
prune: ## Clean up Docker system
	docker system prune -f
	docker volume prune -f
	docker network prune -f