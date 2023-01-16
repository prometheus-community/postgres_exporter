# Ensure that 'all' is the default target otherwise it will be the first target from Makefile.common.
all::

# Needs to be defined before including Makefile.common to auto-generate targets
DOCKER_ARCHS ?= amd64 armv7 arm64 ppc64le
DOCKER_REPO  ?= prometheuscommunity

include Makefile.common

DOCKER_IMAGE_NAME       ?= postgres-exporter

GO_BUILD_LDFLAGS = -X github.com/prometheus/common/version.Version=$(shell cat VERSION) -X github.com/prometheus/common/version.Revision=$(shell git rev-parse HEAD) -X github.com/prometheus/common/version.Branch=$(shell git describe --always --contains --all) -X github.com/prometheus/common/version.BuildUser= -X github.com/prometheus/common/version.BuildDate=$(shell date +%FT%T%z) -s -w

export PMM_RELEASE_PATH?=.

release:
	go build -ldflags="$(GO_BUILD_LDFLAGS)" -o $(PMM_RELEASE_PATH)/postgres_exporter ./cmd/postgres_exporter
