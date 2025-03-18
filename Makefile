# Build and run the app
build-run:
	@echo "Building and running the application with Docker..."
	docker-compose -f docker/docker-compose.yml up --build

# Build the app
build:
	@echo "Building the application with Docker..."
	docker-compose -f docker/docker-compose.yml build

# Run the app
start:
	@echo "Starting the application with Docker..."
	docker-compose -f docker/docker-compose.yml up

# Run the app locally
run:
	# @echo "Starting RabbitMQ with Dockr..."
	# docker-compose -f docker/docker-compose.rabbitmq.yml up -d
	# @sleep 5 # Wait for RabbitMQ to start
	@echo "Running the application locally..."
	go run ./cmd/app/main.go

# Run the app locally with RabbitMQ
run-rmq:
	@echo "Starting RabbitMQ with Dockr..."
	docker-compose -f docker/docker-compose.rabbitmq.yml up -d
	@sleep 5 # Wait for RabbitMQ to start
	@echo "Running the application locally..."
	go run ./cmd/app/main.go

# Run the migrations 
migrate:
	@echo "Running migrations..."
	go run ./cmd/migration_runner/main.go

# Reset the database
reset-db:
	@echo "Resetting the database..."
	@rm -f ./data/app.db
	@echo "Database file deleted."
	@echo "Reapplying migrations..."
	@go run ./cmd/migration_runner/main.go
	@echo "Database reset complete."

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Stop and remove Docker containers
docker-down:
	@echo "Stopping and removing Docker containers..."
	docker-compose -f docker/docker-compose.yml down
	docker-compose -f docker/docker-compose.rabbitmq.yml down

# Clean up Docker resources
docker-clean:
	@echo "Cleaning up Docker resources..."
	docker-compose -f docker/docker-compose.yml down -v --rmi all
	docker-compose -f docker/docker-compose.rabbitmq.yml down -v --rmi all

generate-mocks:
	mockgen -destination=./internal/http/handlers/mocks/mock_repoSvc.go -package=mocks github.com/victor-nach/git-monitor/internal/http/handlers repoSvc
	mockgen -destination=./internal/http/handlers/mocks/mock_taskSvc.go -package=mocks github.com/victor-nach/git-monitor/internal/http/handlers taskSvc
	mockgen -destination=./internal/http/handlers/mocks/mock_commitSvc.go -package=mocks github.com/victor-nach/git-monitor/internal/http/handlers commitSvc