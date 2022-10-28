USER_BALANCE_BINARY=userBalanceApp

up:
	@echo "Starting Docker images..."
	docker-compose up -d
	@echo "Docker images started!"

up_build: build_user-balance-service
	@echo "Stopping docker images (if running...)"
	docker-compose down
	@echo "Building (when required) and starting docker images..."
	docker-compose up --build -d
	@echo "Docker images built and started!"

down:
	@echo "Stopping docker compose..."
	docker-compose down
	@echo "Done!"

build_user-balance-service:
	@echo "Building user balance binary..."
	cd ./../user-balance-service && env GOOS=linux CGO_ENABLED=0 go build -o ${USER_BALANCE_BINARY} ./cmd
	@echo "Done!"