.PHONY: help build test lint clean deps-check deps-update

help:
	@echo "Available targets:"
	@echo "  build         - Build frontend and backend"
	@echo "  test          - Run tests"
	@echo "  lint          - Run linters"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps-check    - Check for outdated dependencies"
	@echo "  deps-update   - Update dependencies"
	@echo "  dev-frontend  - Start frontend dev server"
	@echo "  dev-backend   - Start backend dev server"

build:
	@echo "Building frontend..."
	cd app/ui && npm run build
	@echo "Building backend..."
	cd app/server && go build -o dockpit

test:
	@echo "Running backend tests..."
	cd app/server && go test ./...
	@echo "Running frontend tests..."
	cd app/ui && npm test

lint:
	@echo "Linting backend..."
	cd app/server && golangci-lint run
	@echo "Linting frontend..."
	cd app/ui && npm run lint

clean:
	@echo "Cleaning build artifacts..."
	rm -rf app/ui/dist
	rm -f app/server/dockpit
	rm -f *.fpk

deps-check:
	@bash scripts/check-deps.sh

deps-update:
	@echo "Updating frontend dependencies..."
	cd app/ui && npm update
	@echo "Updating backend dependencies..."
	cd app/server && go get -u ./... && go mod tidy

dev-frontend:
	cd app/ui && npm run dev

dev-backend:
	cd app/server && go run main.go
