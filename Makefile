## variables
GCP_PROJECT_ID = lucasfaria-tools-api
DOCKER_IMAGE_NAME = lucasheriques/go-api
GCP_ARTIFACT_REPO = my-repo

## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

.PHONY: ci
ci:
	go mod tidy
	go mod verify
	go vet ./...
	go test -race -vet=off ./...

## start/docker: start the application with Docker Compose
.PHONY: dev
dev:
	@echo 'Starting application with Docker Compose...'
	docker compose -f docker-compose.yaml -f docker-compose.dev.yaml up --build --remove-orphans

.PHONY: build
build:
	@echo 'Building production application with Docker Compose...'
	docker compose build

.PHONY: start
start:
	@echo 'Starting production application with Docker Compose...'
	docker compose up -d

.PHONY: stop
stop:
	@echo 'Stopping production application with Docker Compose...'
	docker compose down
