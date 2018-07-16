
COVERDIR = .coverage
TOOLDIR = tools
BINDIR = bin
RELEASEDIR = release

DIRS = $(BINDIR) $(RELEASEDIR)

GO_SRC := $(shell find . -name '*.go' ! -path '*/vendor/*' ! -path 'tools/*' ! -path 'bin/*' ! -path 'release/*' )
GO_DIRS := $(shell find . -type d -name '*.go' ! -path '*/vendor/*' ! -path 'tools/*' ! -path 'bin/*' ! -path 'release/*' )
GO_PKGS := $(shell go list ./... | grep -v '/vendor/')

CONTAINER_NAME ?= wrouesnel/postgres_exporter:latest
BINARY := $(shell basename $(shell pwd))
VERSION ?= $(shell git describe --dirty 2>/dev/null)
VERSION_SHORT ?= $(shell git describe --abbrev=0 2>/dev/null)

ifeq ($(VERSION),)
VERSION := v0.0.0
endif

ifeq ($(VERSION_SHORT),)
VERSION_SHORT := v0.0.0
endif

# By default this list is filtered down to some common platforms.
platforms := $(subst /,-,$(shell go tool dist list | grep -e linux -e windows -e darwin | grep -e 386 -e amd64))
PLATFORM_BINS_TMP := $(patsubst %,$(BINDIR)/$(BINARY)_$(VERSION_SHORT)_%/$(BINARY),$(platforms))
PLATFORM_BINS := $(patsubst $(BINDIR)/$(BINARY)_$(VERSION_SHORT)_windows-%/$(BINARY),$(BINDIR)/$(BINARY)_$(VERSION_SHORT)_windows-%/$(BINARY).exe,$(PLATFORM_BINS_TMP))
PLATFORM_DIRS := $(patsubst %,$(BINDIR)/$(BINARY)_$(VERSION_SHORT)_%,$(platforms))
PLATFORM_TARS := $(patsubst %,$(RELEASEDIR)/$(BINARY)_$(VERSION_SHORT)_%.tar.gz,$(platforms))

# These are evaluated on use, and so will have the correct values in the build
# rule (https://vic.demuzere.be/articles/golang-makefile-crosscompile/)
PLATFORMS_TEMP = $(subst /, ,$(subst -, ,$(patsubst $(BINDIR)/$(BINARY)_$(VERSION_SHORT)_%,%,$@)))
GOOS = $(word 1, $(PLATFORMS_TEMP))
GOARCH = $(word 2, $(PLATFORMS_TEMP))

CURRENT_PLATFORM_TMP := $(BINDIR)/$(BINARY)_$(VERSION_SHORT)_$(shell go env GOOS)-$(shell go env GOARCH)/$(BINARY)
CURRENT_PLATFORM := $(patsubst $(BINDIR)/$(BINARY)_$(VERSION_SHORT)_windows-%/$(BINARY),$(BINDIR)/$(BINARY)_$(VERSION_SHORT)_windows-%/$(BINARY).exe,$(CURRENT_PLATFORM_TMP))

CONCURRENT_LINTERS ?=
ifeq ($(CONCURRENT_LINTERS),)
CONCURRENT_LINTERS = $(shell gometalinter --help | grep -o 'concurrency=\w*' | cut -d= -f2 | cut -d' ' -f1)
endif

LINTER_DEADLINE ?= 30s

$(shell mkdir -p $(DIRS))

export PATH := $(TOOLDIR)/bin:$(PATH)
SHELL := env PATH=$(PATH) /bin/bash

all: style lint test binary

binary: $(BINARY)

$(BINARY): $(CURRENT_PLATFORM)
	ln -sf $< $@

$(PLATFORM_BINS): $(GO_SRC)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -a \
		-ldflags "-extldflags '-static' -X main.Version=$(VERSION)" \
		-o $@ .

$(PLATFORM_DIRS): $(PLATFORM_BINS)

$(PLATFORM_TARS): $(RELEASEDIR)/%.tar.gz : $(BINDIR)/%
	tar -czf $@ -C $(BINDIR) $$(basename $<)
	
release-bin: $(PLATFORM_BINS)

release: $(PLATFORM_TARS)

# Take a go build and turn it into a minimal container
docker: $(CURRENT_PLATFORM)
	docker build --build-arg=binary=$(CURRENT_PLATFORM) -t $(CONTAINER_NAME) .

style: tools
	gometalinter --disable-all --enable=gofmt --vendor

lint: tools
	@echo Using $(CONCURRENT_LINTERS) processes
	gometalinter -j $(CONCURRENT_LINTERS) --deadline=$(LINTER_DEADLINE) --disable=gotype --disable=gocyclo $(GO_DIRS)

fmt: tools
	gofmt -s -w $(GO_SRC)

postgres_exporter_integration_test: $(GO_SRC)
	CGO_ENABLED=0 go test -c -tags integration \
		-a -ldflags "-extldflags '-static' -X main.Version=$(VERSION)" \
		-o postgres_exporter_integration_test -cover -covermode count ./collector/...

test: tools
	@mkdir -p $(COVERDIR)
	@rm -f $(COVERDIR)/*
	for pkg in $(GO_PKGS) ; do \
		go test -v -covermode count -coverprofile=$(COVERDIR)/$$(echo $$pkg | tr '/' '-').out $$pkg || exit 1 ; \
	done
	gocovmerge $(shell find $(COVERDIR) -name '*.out') > cover.test.out

test-integration: postgres_exporter postgres_exporter_integration_test
	tests/test-smoke "$(shell pwd)/postgres_exporter" "$(shell pwd)/postgres_exporter_integration_test_script $(shell pwd)/postgres_exporter_integration_test $(shell pwd)/cover.integration.out"

cover.out: tools
	gocovmerge cover.*.out > cover.out

clean:
	[ ! -z $(BINDIR) ] && [ -e $(BINDIR) ] && find $(BINDIR) -print -delete || /bin/true
	[ ! -z $(COVERDIR) ] && [ -e $(COVERDIR) ] && find $(COVERDIR) -print -delete || /bin/true
	[ ! -z $(RELEASEDIR) ] && [ -e $(RELEASEDIR) ] && find $(RELEASEDIR) -print -delete || /bin/true
	rm -f postgres_exporter postgres_exporter_integration_test
	
tools:
	$(MAKE) -C $(TOOLDIR)
	
.PHONY: tools style fmt test all release binary clean
