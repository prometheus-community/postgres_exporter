GO             ?= GO15VENDOREXPERIMENT=1 go
GO_SRC         := $(shell find -type f -name '*.go' ! -path '*/vendor/*')

CONTAINER_NAME ?= wrouesnel/postgres_exporter:latest
MEGACHECK      ?= $(GOPATH)/bin/megacheck
VERSION        ?= $(shell git describe --dirty)
pkgs            = $(shell $(GO) list ./... | grep -v /vendor/)
TARGET         ?= postgres_exporter

all: fmt vet megacheck test postgres_exporter

# Cross compilation (e.g. if you are on a Mac)
cross: docker-build docker

# Simple go build
postgres_exporter: $(GO_SRC)
	@echo ">> building binaries"
	@CGO_ENABLED=0;$(GO) build -a -ldflags "-extldflags '-static' -X main.Version=$(VERSION)" -o $(TARGET) .

postgres_exporter_integration_test: $(GO_SRC)
	@CGO_ENABLED=0;$(GO) test -c -tags integration \
	    -a -ldflags "-extldflags '-static' -X main.Version=$(VERSION)" -o postgres_exporter_integration_test -cover -covermode count .

# Take a go build and turn it into a minimal container
docker: postgres_exporter
	docker build -t $(CONTAINER_NAME) .

vet:
	@echo ">> vetting code"
	@$(GO) vet $(pkgs)

# Check code conforms to go fmt
style:
	! gofmt -s -l $(GO_SRC) 2>&1 | read 2>/dev/null

# Format the code
fmt:
	@echo ">> formatting code"
	@$(GO) fmt $(pkgs)

megacheck: $(MEGACHECK)
	@echo ">> megacheck code"
	@$(MEGACHECK) $(pkgs)

test:
	@echo ">> running tests"
	@$(GO) test -v -covermode count -coverprofile=cover.test.out $(pkgs)

test-integration: postgres_exporter postgres_exporter_integration_test
	tests/test-smoke "$(shell pwd)/postgres_exporter" "$(shell pwd)/postgres_exporter_integration_test_script $(shell pwd)/postgres_exporter_integration_test $(shell pwd)/cover.integration.out"

# Do a self-contained docker build - we pull the official upstream container
# and do a self-contained build.
docker-build:
	docker run -v $(shell pwd):/go/src/github.com/wrouesnel/postgres_exporter \
	    -v $(shell pwd):/real_src \
	    -e SHELL_UID=$(shell id -u) -e SHELL_GID=$(shell id -g) \
	    -w /go/src/github.com/wrouesnel/postgres_exporter \
		golang:1.7-wheezy \
		/bin/bash -c "make >&2 && chown $$SHELL_UID:$$SHELL_GID ./postgres_exporter"
	docker build -t $(CONTAINER_NAME) .

$(GOPATH)/bin/megacheck mega:
	@GOOS=$(shell uname -s | tr A-Z a-z) \
		GOARCH=$(subst x86_64,amd64,$(patsubst i%86,386,$(shell uname -m))) \
		$(GO) get -u honnef.co/go/tools/cmd/megacheck

push:
	docker push $(CONTAINER_NAME)

clean:
	@echo ">> Cleaning up"
	@rm -f $(TARGET) *~

.PHONY: clean docker-build docker test vet push cross $(GOPATH)/bin/megacheck mega
