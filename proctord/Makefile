#!make
include .env.test
export $(shell sed 's/=.*//' .env.test)

APP=proctord
APP_EXECUTABLE=./$(APP)

build-deps:
	glide install

build:
	go build -o $(APP)

test:
	go test -race -cover $(shell glide novendor)

test-verbose:
	go test -race -cover $(shell glide novendor) -v

db.setup: db.create db.migrate

db.create:
	PGPASSWORD=$(PROCTOR_POSTGRES_PASSWORD) psql -h $(PROCTOR_POSTGRES_HOST) -p $(PROCTOR_POSTGRES_PORT) -c 'create database $(PROCTOR_POSTGRES_DATABASE);' -U $(PROCTOR_POSTGRES_USER)

db.migrate: $(APP_EXECUTABLE)
	$(APP_EXECUTABLE) migrate

db.teardown:
	-PGPASSWORD=$(PROCTOR_POSTGRES_PASSWORD) psql -h $(PROCTOR_POSTGRES_HOST) -p $(PROCTOR_POSTGRES_PORT) -c 'drop database $(PROCTOR_POSTGRES_DATABASE);' -U $(PROCTOR_POSTGRES_USER)

setup-test-teardown: build-deps build db.teardown db.setup test-verbose db.teardown

setup-test: build-deps build db.teardown db.setup test
