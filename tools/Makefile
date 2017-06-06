# Makefile to build the tools used in the build system.
# If recreating from scratch, you will need a local install of govendor
# and to run govendor init in this folder before running govendor fetch.

# Ensure we use local bin dir
export PATH := bin:$(PATH)
SHELL := env PATH=$(PATH) /bin/bash

THIS_FILE := $(lastword $(MAKEFILE_LIST))

# This function is used to get the linters used by metalinter
get_metalinters := gometalinter --help | grep -oP '  \w+  \(.+\)' | tr -s ' ' | cut -d' ' -f3 | grep -oP '[^()]+'

TOOL_SRCS := github.com/kardianos/govendor \
 github.com/wadey/gocovmerge \
 github.com/mattn/goveralls \
 github.com/alecthomas/gometalinter

METATOOL_SRCS := 

GO_SRC := $(shell find $(SOURCEDIR) -name '*.go')

GO := GOPATH=$(shell pwd) go

DEFAULT: all

tools.deps: $(GO_SRC)
	# Generate build patterns for static tools
	for pkg in $(TOOL_SRCS); do \
		echo -e "bin/$$(basename $$pkg): $$GO_SRC\n\t\$$(GO) install -v $$pkg" ; \
	done > tools.deps

metatools.deps: tools.deps bin/gometalinter $(GO_SRC)
	# Generate build patterns for metalinters tools
	echo -e "METATOOL_SRCS+=$(shell $(get_metalinters))" > metatools.deps
	for pkg in $(shell $(get_metalinters)) ; do \
		echo -e "bin/$$(basename $$pkg): $$GO_SRC\n\t\$$(GO) install -v $$pkg" ; \
	done >> metatools.deps

include tools.deps
include metatools.deps
	
update:
	# Fetch govendor, then rebuild govendor.
	govendor fetch github.com/kardianos/govendor
	$(GO) install -v github.com/kardianos/govendor
	# Fetch gometalinter and rebuild gometalinter.
	govendor fetch github.com/alecthomas/gometalinter
	$(GO) install -v github.com/alecthomas/gometalinter
	$(MAKE) -f $(THIS_FILE) update-phase-2

update-phase-2:
	# Fetch the new metalinter list.
	for pkg in $(TOOL_SRCS) $$($(get_metalinters)); do \
		govendor fetch -v $$pkg ; \
	done	

clean:
	rm -rf bin pkg tools.deps metatools.deps

all: $(addprefix bin/,$(notdir $(TOOL_SRCS) $(METATOOL_SRCS) ))

# TOOL_SRCS is included here since we'll never really have these files.
.PHONY: all update clean $(TOOL_SRCS) $(METATOOL_SRCS)
