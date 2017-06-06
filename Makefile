
GO_SRC := $(shell find -type f -name '*.go' ! -path '*/vendor/*')

CONTAINER_NAME ?= k8sdb/postgres_exporter:latest
VERSION ?= $(shell git describe --dirty)

all: vet test postgres_exporter

# Cross compilation (e.g. if you are on a Mac)
cross: docker-build docker

# Simple go build
postgres_exporter: $(GO_SRC)
	CGO_ENABLED=0 go build -a -ldflags "-extldflags '-static' -X main.Version=$(VERSION)" -o postgres_exporter .

postgres_exporter_integration_test: $(GO_SRC)
	CGO_ENABLED=0 go test -c -tags integration \
	    -a -ldflags "-extldflags '-static' -X main.Version=$(VERSION)" -o postgres_exporter_integration_test -cover -covermode count .

# Take a go build and turn it into a minimal container
docker: postgres_exporter
	docker build -t $(CONTAINER_NAME) .

vet:
	go vet

# Check code conforms to go fmt
style:
	! gofmt -s -l $(GO_SRC) 2>&1 | read 2>/dev/null

# Format the code
fmt:
	gofmt -s -w $(GO_SRC)

test:
	go test -v -covermode count -coverprofile=cover.test.out

test-integration: postgres_exporter postgres_exporter_integration_test
	tests/test-smoke "$(shell pwd)/postgres_exporter" "$(shell pwd)/postgres_exporter_integration_test_script $(shell pwd)/postgres_exporter_integration_test $(shell pwd)/cover.integration.out"

# Do a self-contained docker build - we pull the official upstream container
# and do a self-contained build.
docker-build:
	docker run -v $(shell pwd):/go/src/github.com/k8sdb/postgres_exporter \
	    -v $(shell pwd):/real_src \
	    -e SHELL_UID=$(shell id -u) -e SHELL_GID=$(shell id -g) \
	    -w /go/src/github.com/k8sdb/postgres_exporter \
		golang:1.7-wheezy \
		/bin/bash -c "make >&2 && chown $$SHELL_UID:$$SHELL_GID ./postgres_exporter"
	docker build -t $(CONTAINER_NAME) .

push:
	docker push $(CONTAINER_NAME)

.PHONY: docker-build docker test vet push cross
