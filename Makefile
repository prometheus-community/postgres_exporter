# Copyright 2018 The Prometheus Authors
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

GO           ?= go
GOFMT        ?= $(GO)fmt
FIRST_GOPATH := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
PROMU        := bin/promu
STATICCHECK  := bin/staticcheck
pkgs          = ./...

PREFIX                  ?= $(shell pwd)
BIN_DIR                 ?= $(shell pwd)
DOCKER_IMAGE_TAG        ?= $(subst /,-,$(shell git rev-parse --abbrev-ref HEAD))
DOCKER_IMAGE_NAME ?= postgres-exporter

all: vet style staticcheck build test

style:
	@echo ">> checking code style"
	! $(GOFMT) -d $$(find . -name '*.go' -print) | grep '^'

check_license:
	@echo ">> checking license header"
	@licRes=$$(for file in $$(find . -type f -iname '*.go') ; do \
               awk 'NR<=3' $$file | grep -Eq "(Copyright|generated|GENERATED)" || echo $$file; \
       done); \
       if [ -n "$${licRes}" ]; then \
               echo "license header checking failed:"; echo "$${licRes}"; \
               exit 1; \
       fi

test-short:
	@echo ">> running short tests"
	$(GO) test -short $(pkgs)

test:
	@echo ">> running all tests"
	$(GO) test -race $(pkgs)

format:
	@echo ">> formatting code"
	$(GO) fmt $(pkgs)

vet:
	@echo ">> vetting code"
	$(GO) vet $(pkgs)

staticcheck: $(STATICCHECK)
	@echo ">> running staticcheck"
	GOOS= GOARCH= $(GO) build -modfile=tools/go.mod -o bin/staticcheck honnef.co/go/tools/cmd/staticcheck
	bin/staticcheck $(pkgs)

build: promu
	@echo ">> building binaries"
	$(PROMU) build --prefix $(PREFIX)

tarball: promu
	@echo ">> building release tarball"
	unexport GOBIN
	$(PROMU) tarball --prefix $(PREFIX) $(BIN_DIR)

docker:
	docker build -t "$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)" .

promu:
	GOOS= GOARCH= $(GO) build -modfile=tools/go.mod -o bin/promu github.com/prometheus/promu

.PHONY: all style check_license format build test vet assets tarball docker promu staticcheck bin/staticcheck
