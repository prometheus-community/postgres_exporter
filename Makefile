
GO_SRC := $(shell find . -type f -name "*.go")

CONTAINER_NAME ?= wrouesnel/postgres_exporter:latest

all: vet test postgres_exporter

# Simple go build
postgres_exporter: $(GO_SRC)
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags "-extldflags '-static' -X main.Version=git:$(shell git rev-parse HEAD)" -o postgres_exporter .

# Take a go build and turn it into a minimal container
docker: postgres_exporter
	tar -cf - postgres_exporter | docker import --change "EXPOSE 9113" \
			--change 'ENTRYPOINT [ "/postgres_exporter" ]' \
			- $(CONTAINER_NAME)

vet:
	go vet .

test:
	go test -v .

test-integration:
	tests/test-smoke

# Do a self-contained docker build - we pull the official upstream container,
# then template out a dockerfile which builds the real image.
docker-build: postgres_exporter
	docker run -v $(shell pwd):/go/src/github.com/wrouesnel/postgres_exporter \
	    -w /go/src/github.com/wrouesnel/postgres_exporter \
		golang:1.6-wheezy \
		/bin/bash -c "make >&2 && tar -cf - ./postgres_exporter" | \
		docker import --change "EXPOSE 9113" \
			--change 'ENTRYPOINT [ "/postgres_exporter" ]' \
			- $(CONTAINER_NAME)

.PHONY: docker-build docker test vet
