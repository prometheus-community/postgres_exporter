# Ensure that 'all' is the default target otherwise it will be the first target from Makefile.common.
all::

# Needs to be defined before including Makefile.common to auto-generate targets
DOCKER_ARCHS ?= amd64 armv7 arm64 ppc64le
DOCKER_REPO  ?= prometheuscommunity

include Makefile.common

DOCKER_IMAGE_NAME       ?= postgres-exporter

CHLOGGEN = GOOS= GOARCH= $(GO) run go.opentelemetry.io/build-tools/chloggen@v0.30.0
CHLOGGEN_CONFIG ?= .chloggen/config.yaml

.PHONY: chlog-new chlog-validate chlog-preview chlog-update
chlog-new:
	@test -n "$(NAME)" || (echo 'Usage: make chlog-new NAME=<short-name>'; exit 1)
	$(CHLOGGEN) --config "$(CHLOGGEN_CONFIG)" new --filename "$(NAME)"

chlog-validate:
	$(CHLOGGEN) --config "$(CHLOGGEN_CONFIG)" validate

chlog-preview: chlog-validate
	$(CHLOGGEN) --config "$(CHLOGGEN_CONFIG)" update --dry

chlog-update: chlog-validate
	@test -n "$(VERSION)" || (echo 'Usage: make chlog-update VERSION="0.x.0 / YYYY-MM-DD"'; exit 1)
	$(CHLOGGEN) --config "$(CHLOGGEN_CONFIG)" update --version "$(VERSION)"
