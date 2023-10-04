$(shell (if [ ! -e .env ]; then cp default.env .env; fi))
include .env
export

RUN_ARGS = $(filter-out $@,$(MAKECMDGOALS))

include .make/utils.mk
include .make/docker-compose-shared-services.mk

.PHONY: install
install: erase build start-all wait## clean current environment, recreate dependencies and spin up again;

.PHONY: install-lite
install-lite: build start

.PHONY: start
start: ##up-services ## spin up environment
	docker-compose up -d

.PHONY: stop
stop: ## stop environment
	docker-compose stop

.PHONY: start services
start-services: shared-service-up ## up shared services

.PHONY: stop services
stop-services: shared-service-stop ## stop shared services

.PHONY: start-all
start-all: start start-services ## start full project environment

.PHONY: stop-all
stop-all: stop stop-services ## stop full project environment

.PHONY: remove
remove: ## remove project docker containers
	docker-compose rm -v -f

.PHONY: erase
erase: stop-all remove shared-service-erase docker-remove-volumes ## stop and delete containers, clean volumes

.PHONY: build
build: ## build environment and initialize composer and project dependencies
	docker build .docker/go/ -t docker-gs-ping
	docker-compose build --pull

.PHONY: setup
setup: setup-enqueue ## setup-db build environment and initialize composer and project dependencies

.PHONY: clean
clean: ## Clear build vendor report folders
	rm -rf build/ vendor/ var/

.PHONY: migrate
migrate: ## run migrations
	migrate -database $(DB_URL) -path /migrations up

.PHONY: release
release:
	docker build -f Dockerfile -t docker-rest-events --build-arg GITHUB_SHA=$(GITHUB_SHA) .
	docker rm -f docker-rest-events 2>/dev/null
	docker run -d -p 127.0.0.1:8080:8080/tcp --env POSTGRES_DSN='$(POSTGRES_DSN)' docker-rest-events

.PHONY: build-app
build-app:
	go build  -ldflags "-X main.revision=$(RUN_ARGS) -s -w" app/backend/main.go
	mv main rest-events
