ifeq ($(OS),Windows_NT)
    CWD := $(lastword $(dir $(realpath $(MAKEFILE_LIST)/../)))
else
    CWD := $(abspath $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST))))/../))/
endif

shared-service-up:
	docker-compose --project-directory $(CWD)/ -f $(CWD)/docker-compose-shared-services.yml up -d

shared-service-erase:
	docker-compose --project-directory $(CWD)/ -f $(CWD)/docker-compose-shared-services.yml stop
	docker-compose --project-directory $(CWD)/ -f $(CWD)/docker-compose-shared-services.yml down -v --remove-orphans

shared-service-stop:
	docker-compose --project-directory $(CWD)/ -f $(CWD)/docker-compose-shared-services.yml stop

shared-service-logs:
	docker-compose --project-directory $(CWD)/ -f $(CWD)/docker-compose-shared-services.yml logs -f

shared-service-setup-db:
	docker-compose --project-directory $(CWD)/ -f $(CWD)/docker-compose-shared-services.yml exec postgres bash -c "if PGPASSWORD=$(POSTGRES_PASSWORD) psql -U $(POSTGRES_USER) -w -lqtA | cut -d \| -f 1 | grep $(POSTGRESQL_DB); then echo DB $(POSTGRESQL_DB) already exists; else PGPASSWORD=$(POSTGRES_PASSWORD) createdb -U $(POSTGRES_USER) -w $(POSTGRESQL_DB); fi"
