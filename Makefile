GO    := GO15VENDOREXPERIMENT=1 go
PROMU := $(GOPATH)/bin/promu
pkgs   = $(shell $(GO) list ./... | grep -v /vendor/)
PREFIX              ?= $(shell pwd)
BIN_DIR             ?= $(shell pwd)

CONTAINER_NAME ?= wrouesnel/postgres_exporter:latest
DOCKER_IMAGE_TAG    ?= $(subst /,-,$(shell git rev-parse --abbrev-ref HEAD))

all: format vet build
style:
	@echo ">> checking code style"
	@! gofmt -d $(shell find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

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


docker: postgres_exporter
	@echo ">> building docker image"
	docker run -v $(shell pwd):/go/src/github.com/wrouesnel/postgres_exporter \
	    -w /go/src/github.com/wrouesnel/postgres_exporter \
		golang:1.6-wheezy \
		/bin/bash -c "make >&2 && tar -cf - ./postgres_exporter" | \
		docker import --change "EXPOSE 9113" \
			--change 'ENTRYPOINT [ "/postgres_exporter" ]' \
			- $(CONTAINER_NAME)

promu:
		@GOOS=$(shell uname -s | tr A-Z a-z) \
		GOARCH=$(subst x86_64,amd64,$(patsubst i%86,386,$(shell uname -m))) \
		$(GO) get -u github.com/prometheus/promu


.PHONY: all style format build test vet tarball docker promu
test:
	tests/test-smoke


.PHONY: docker-build docker test vet
