# Ensure that 'all' is the default target otherwise it will be the first target from Makefile.common.
all::

# Needs to be defined before including Makefile.common to auto-generate targets
DOCKER_REPO  ?= form3tech
CROSS_BUILD_PROMUOPTS := -p linux/arm64 #-p linux/amd64 -p windows/amd64

include Makefile.common

DOCKER_IMAGE_NAME       ?= postgres-exporter
