APP=proctord
DB_USER=$(PROCTOR_POSTGRES_USER)
DB_NAME=$(PROCTOR_POSTGRES_DATABASE)
APP_EXECUTABLE=./$(APP)

build-deps:
	glide install

build: build-deps
	go build

test:
	source .env.test && go test $(shell glide novendor)

db.setup: db.create db.migrate

db.create:
	createdb -O$(DB_USER) -Eutf8 $(DB_NAME)

db.migrate: $(APP_EXECUTABLE)
	$(APP_EXECUTABLE) migrate

copy-config:
	cp .env.sample .env
