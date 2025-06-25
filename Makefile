.PHONY: help dev prod build-dev build-prod up-dev up-prod down clean logs test

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

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

# General targets
down: ## Stop all containers
	docker-compose down
	docker-compose -f docker-compose.prod.yml down

clean: ## Remove all containers and images
	docker-compose down --rmi all --volumes --remove-orphans
	docker-compose -f docker-compose.prod.yml down --rmi all --volumes --remove-orphans

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