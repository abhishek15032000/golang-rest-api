APP_NAME=rest-api
BIN_NAME=gobin
BUILD_DIR=./bin
GO_FILES=$(shell find . -name '*.go' -not -path "./vendor/*")

run:
	@echo "🚀 🚀 🚀 🚀 Running the server ..."
	@go run main.go || true

deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@echo "Dependencies installed successfully!"

fmt:
	@echo "Formatting code ..."
	@go fmt ./...

build:
	@echo "Building the project ..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BIN_NAME) main.go
	@echo "Build completed successfully! $(BUILD_DIR)"

clean:
	@echo "Cleaning up ..."
	@rm -rf $(BUILD_DIR)
	@echo "Cleaned up successfully!"

stop:
	@echo "Stopping server...."
	@pkill -f "go run main.go" || echo "no process found"

migrate-up:
	@echo "Running database migration"
	@dbmate --env-file .env -d ./db/migrations up
	@echo "Migration applied successfully"

migrate-down:
	@echo "Rolling back last migration"
	@dbmate --env-file .env -d ./db/migrations down
	@echo "Migration rolled back successfully"

migrate-status:
	@dbmate --env-file .env -d ./db/migrations status

migrate-reset:
	@echo "Rolling back ALL migrations"
	@while dbmate --env-file .env -d ./db/migrations status 2>&1 | grep -q "\[X\]"; do \
		dbmate --env-file .env -d ./db/migrations down; \
	done
	@echo "All migrations rolled back"


help:
	@echo "Available commands"
	@echo " make run           - Run the server"
	@echo " make deps          - Install dependencies"
	@echo " make fmt           - Format the code"
	@echo " make build         - Build the project"
	@echo " make clean         - Clean up the binary"
	@echo " make stop          - Stop all running server"
	@echo " make migrate-up    - Run database migrations"
	@echo " make migrate-down  - Roll back migrations"