.PHONY: all
all: build-deps build fmt vet lint test

GLIDE_NOVENDOR=$(shell glide novendor | grep -v proctord)

setup:
	mkdir -p $(GOPATH)/bin
	go get -u github.com/golang/lint/golint
	which glide &>/dev/null
	if [ $? -ne 0 ]; then curl https://glide.sh/get | sh; fi

build-deps:
	glide install

update-deps:
	glide update

compile:
	mkdir -p out/
	go build -race $(GLIDE_NOVENDOR)

build: build-deps compile fmt vet lint

fmt:
	go fmt $(GLIDE_NOVENDOR)

vet:
	go vet $(GLIDE_NOVENDOR)

lint:
	@for p in $(UNIT_TEST_PACKAGES); do \
		echo "==> Linting $$p"; \
		golint -set_exit_status $$p; \
	done

install: setup build-deps compile
	go install
