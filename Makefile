GO    := GO15VENDOREXPERIMENT=1 go
PROMU := $(GOPATH)/bin/promu
pkgs   = $(shell $(GO) list ./... | grep -v /vendor/)
PREFIX              ?= $(shell pwd)
BIN_DIR             ?= $(shell pwd)

GO_SRC := $(shell find -type f -name "*.go")

CONTAINER_NAME ?= wrouesnel/postgres_exporter:latest
DOCKER_IMAGE_TAG    ?= $(subst /,-,$(shell git rev-parse --abbrev-ref HEAD))

all: format vet build test 
style:
	@echo ">> checking code style"
	@! gofmt -d $(shell find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

test:
		@echo ">> running tests"
		@$(GO) test -v .

format:
		@echo ">> formatting code"
			@$(GO) fmt $(pkgs)
	
vet:
	@echo ">> vetting code"
	@$(GO) vet $(pkgs)

build: promu
		@echo ">> building binaries"
			@$(PROMU) build --prefix $(PREFIX)

tarball: promu
		@echo ">> building release tarball"
			@$(PROMU) tarball --prefix $(PREFIX) $(BIN_DIR)

docker:
		@echo ">> building docker image"
			@docker build -t "$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)" .

promu:
		@GOOS=$(shell uname -s | tr A-Z a-z) \
					GOARCH=$(subst x86_64,amd64,$(patsubst i%86,386,$(shell uname -m))) \
							$(GO) get -u github.com/prometheus/promu


.PHONY: all style format build test vet tarball docker promu
test-integration:
	tests/test-smoke


.PHONY: docker-build docker test vet
