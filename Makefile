# Fix Track Bot Makefile

.PHONY: help build run test clean docker-build docker-up docker-down docker-logs

# Default target
help:
	@echo "Available commands:"
	@echo "  build         - Build the Go application"
	@echo "  run          - Run the application locally"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-up    - Start services with Docker Compose"
	@echo "  docker-down  - Stop services with Docker Compose"
	@echo "  docker-logs  - View Docker Compose logs"
	@echo "  docker-restart - Restart all services"

# Build the Go application
build:
	go build -o fix-track-bot .

# Run the application locally
run:
	go run .

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -f fix-track-bot
	go clean

# Docker commands
docker-build:
	docker-compose build

docker-up:
	@echo "Starting PostgreSQL and Fix Track Bot..."
	@echo "Make sure to copy docker.env.example to .env and set your DISCORD_TOKEN"
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-restart:
	docker-compose restart

# Setup for first time
setup:
	@echo "Setting up Fix Track Bot..."
	@echo "1. Copying environment file..."
	@cp docker.env.example .env
	@echo "2. Please edit .env file and set your DISCORD_TOKEN"
	@echo "3. Run 'make docker-up' to start the services"

# Database commands
db-reset:
	docker-compose down -v
	docker-compose up -d postgres
	sleep 10
	docker-compose up -d fix-track-bot
