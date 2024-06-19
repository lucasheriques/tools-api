## help: print this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	tilt up

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	go build -ldflags='-s' -o=./bin/api ./cmd/api

## start/api: start the cmd/api application
.PHONY: start/api
start/api:
	@echo 'Starting cmd/api...'
	./bin/api

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


## variables
GCP_PROJECT_ID = lucasfaria-tools-api
DOCKER_IMAGE_NAME = tools-lucasfaria-dev
GCP_ARTIFACT_REPO = my-repo

## docker/build: builds the docker container image
.PHONY: docker/build
docker/build:
	@echo 'Building docker image...'
	docker build -t ${DOCKER_IMAGE_NAME}:latest --file ./deployments/Dockerfile .

.PHONY: docker/build-and-push
docker/build-and-push:
	@echo 'Building docker image...'
	docker build -t ${DOCKER_IMAGE_NAME}:latest --file ./deployments/Dockerfile .
	@echo 'Pushing docker image to GCP...'
	docker tag ${DOCKER_IMAGE_NAME}:latest us-central1-docker.pkg.dev/${GCP_PROJECT_ID}/${GCP_ARTIFACT_REPO}/${DOCKER_IMAGE_NAME}:latest
	docker push us-central1-docker.pkg.dev/${GCP_PROJECT_ID}/${GCP_ARTIFACT_REPO}/${DOCKER_IMAGE_NAME}:latest

.PHONY: terraform
terraform:
	@echo 'Running terraform...'
	terraform init
	terraform plan
	terraform apply


.PHONY: tf/apply
tf/apply:
	@echo 'Running terraform apply...'
	cd terraform
	terraform apply


## start/docker: start the application with Docker Compose
.PHONY: dev
dev:
	@echo 'Starting application with Docker Compose...'
	docker-compose up --build

## stop/docker: stop the application with Docker Compose
.PHONY: stop/docker
stop/docker:
	@echo 'Stopping application with Docker Compose...'
	docker-compose down

## restart/docker: restart the application with Docker Compose
.PHONY: restart/docker
restart/docker:
	@echo 'Restarting application with Docker Compose...'
	docker-compose down
	docker-compose up --build
