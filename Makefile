SHELL := /bin/bash

#!make

include .env.test
export $(shell sed 's/=.*//' .env.test)

.EXPORT_ALL_VARIABLES:
SRC_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
OUT_DIR := $(SRC_DIR)/_output
BIN_DIR := $(OUT_DIR)/bin
PLUGIN_DIR := $(BIN_DIR)/plugin
FTEST_DIR := test/procs
CONFIG_DIR := test/config
GOPROXY ?= https://proxy.golang.org
GO111MODULE := on

$(@info $(shell mkdir -p $(OUT_DIR) $(BIN_DIR) $(PLUGIN_DIR))

.PHONY: build
build: test-with-race server cli

.PHONY: test-with-race
test-with-race: RACE_FLAG = -race
test-with-race: test

.PHONY: test
test:
	ENABLE_INTEGRATION_TEST=false \
	go test -race -coverprofile=$(OUT_DIR)/coverage.out ./...

.PHONY: itest
itest: plugin.auth
	PROCTOR_AUTH_PLUGIN_BINARY=$(PLUGIN_DIR)/auth.so \
	ENABLE_INTEGRATION_TEST=true \
	go test -p 1 -race -coverprofile=$(OUT_DIR)/coverage.out ./...

.PHONY: plugin.itest
plugin.itest:
	ENABLE_PLUGIN_INTEGRATION_TEST=true \
	go test -p 1 -race -coverprofile=$(OUT_DIR)/coverage.out ./plugins/...

.PHONY: server
server:
	go build -race -o $(BIN_DIR)/server ./cmd/server/main.go

.PHONY: plugin.auth
plugin.auth:
	go build -race -buildmode=plugin -o $(PLUGIN_DIR)/auth.so ./plugins/gate-auth-plugin/auth.go

.PHONY: cli
cli:
	go build -race -o $(BIN_DIR)/cli ./cmd/cli/main.go

build-all: server cli plugin.auth

.PHONY: start-server
start-server:
	$(BIN_DIR)/server s

generate:
	go get -u github.com/go-bindata/go-bindata/...
	$(GOPATH)/bin/go-bindata -pkg config -o internal/app/cli/config/data.go internal/app/cli/config_template.yaml

db.setup: db.create db.migrate

db.create:
	PGPASSWORD=$(PROCTOR_POSTGRES_PASSWORD) psql -h $(PROCTOR_POSTGRES_HOST) -p $(PROCTOR_POSTGRES_PORT) -c 'create database $(PROCTOR_POSTGRES_DATABASE);' -U $(PROCTOR_POSTGRES_USER)

db.migrate: server
	$(BIN_DIR)/server migrate

db.rollback: server
	$(BIN_DIR)/server rollback

db.teardown:
	-PGPASSWORD=$(PROCTOR_POSTGRES_PASSWORD) psql -h $(PROCTOR_POSTGRES_HOST) -p $(PROCTOR_POSTGRES_PORT) -c 'drop database $(PROCTOR_POSTGRES_DATABASE);' -U $(PROCTOR_POSTGRES_USER)
	redis-cli FLUSHALL

.PHONY: ftest.package.procs
ftest.package.procs:
	PROCTOR_JOBS_PATH=$(FTEST_DIR) \
	ruby ./test/package_procs.rb

.PHONY: ftest.update.metadata
ftest.update.metadata:
	PROCTOR_JOBS_PATH=$(FTEST_DIR) \
	PROCTOR_URI=http://localhost:$(PROCTOR_APP_PORT)/metadata \
	ruby ./test/update_metadata.rb

.PHONY: ftest.update.secret
ftest.update.secret:
	curl -X POST \
	  http://localhost:5000/secret \
	  -H 'Content-Type: application/json' \
	  -d '{"job_name": "say-hello-world","secrets": {"SAMPLE_SECRET_ONE": "Secret One :*","SAMPLE_SECRET_TWO": "Secret Two :V"}}'

.PHONY: ftest.proctor.list
ftest.proctor.list:
	LOCAL_CONFIG_DIR=$(CONFIG_DIR) $(BIN_DIR)/cli list

.PHONY: ftest.proctor.describe
ftest.proctor.describe:
	LOCAL_CONFIG_DIR=$(CONFIG_DIR) $(BIN_DIR)/cli describe say-hello-world

.PHONY: ftest.proctor.execute
ftest.proctor.execute:
	LOCAL_CONFIG_DIR=$(CONFIG_DIR) $(BIN_DIR)/cli execute say-hello-world SAMPLE_ARG_ONE=foo SAMPLE_ARG_TWO=bar

.PHONY: ftest.proctor.execute-with-yaml
ftest.proctor.execute-with-yaml:
	LOCAL_CONFIG_DIR=$(CONFIG_DIR) $(BIN_DIR)/cli execute say-hello-world -f $(FTEST_DIR)/say-hello-world/say_hello_world.yaml

.PHONY: ftest.proctor.logs
ftest.proctor.logs:
	LOCAL_CONFIG_DIR=$(CONFIG_DIR) $(BIN_DIR)/cli logs $(EXECUTION_ID)

.PHONY: ftest.proctor.status
ftest.proctor.status:
	LOCAL_CONFIG_DIR=$(CONFIG_DIR) $(BIN_DIR)/cli status $(EXECUTION_ID)
