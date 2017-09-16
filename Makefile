
COVERDIR = .coverage
TOOLDIR = tools

GO_SRC := $(shell find . -name '*.go' ! -path '*/vendor/*' ! -path 'tools/*' )
GO_DIRS := $(shell find . -type d -name '*.go' ! -path '*/vendor/*' ! -path 'tools/*' )
GO_PKGS := $(shell go list ./... | grep -v '/vendor/')

CONTAINER_NAME ?= wrouesnel/postgres_exporter:latest
VERSION ?= $(shell git describe --dirty)

CONCURRENT_LINTERS ?= $(shell cat /proc/cpuinfo | grep processor | wc -l)
LINTER_DEADLINE ?= 30s

export PATH := $(TOOLDIR)/bin:$(PATH)
SHELL := env PATH=$(PATH) /bin/bash
# We may want to move shell scripts to have a common *.sh suffix, 
# allowing us to use 'find *.sh', like we do for GO_SRC.
hash := \#
SHELL_SRC := $(shell grep -l '$(hash)!/bin/bash' -r . \
                          --exclude-dir 'tools' \
                          --exclude Makefile)

all: style lint shellcheck test postgres_exporter

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

style: tools
	gometalinter --disable-all --enable=gofmt --vendor

shellcheck:
	shellcheck $(SHELL_SRC)

lint: tools
	@echo Using $(CONCURRENT_LINTERS) processes
	gometalinter -j $(CONCURRENT_LINTERS) --deadline=$(LINTER_DEADLINE) --disable=gotype --disable=gocyclo $(GO_DIRS)

fmt: tools
	gofmt -s -w $(GO_SRC)

test: tools
	@mkdir -p $(COVERDIR)
	@rm -f $(COVERDIR)/*
	for pkg in $(GO_PKGS) ; do \
		go test -v -covermode count -coverprofile=$(COVERDIR)/$$(echo $$pkg | tr '/' '-').out $$pkg ; \
	done
	gocovmerge $(shell find $(COVERDIR) -name '*.out') > cover.test.out

test-integration: postgres_exporter postgres_exporter_integration_test
	tests/test-smoke "$(shell pwd)/postgres_exporter" "$(shell pwd)/postgres_exporter_integration_test_script $(shell pwd)/postgres_exporter_integration_test $(shell pwd)/cover.integration.out"

cover.out: tools
	gocovmerge cover.*.out > cover.out

# Do a self-contained docker build - we pull the official upstream container
# and do a self-contained build.
docker-build:
	docker run -v $(shell pwd):/go/src/github.com/wrouesnel/postgres_exporter \
	    -v $(shell pwd):/real_src \
	    -e SHELL_UID=$(shell id -u) -e SHELL_GID=$(shell id -g) \
	    -w /go/src/github.com/wrouesnel/postgres_exporter \
		golang:1.8-wheezy \
		/bin/bash -c "make >&2 && chown $$SHELL_UID:$$SHELL_GID ./postgres_exporter"
	docker build -t $(CONTAINER_NAME) .

push:
	docker push $(CONTAINER_NAME)

tools:
	$(MAKE) -C $(TOOLDIR)

clean:
	rm -f postgres_exporter postgres_exporter_integration_test

.PHONY: tools docker-build docker lint shellcheck fmt test vet push cross clean
