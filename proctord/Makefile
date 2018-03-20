build-deps:
	glide install

build: build-deps
	go build

test:
	source .env.test && go test $(shell glide novendor)
