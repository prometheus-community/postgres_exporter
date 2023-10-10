PLATFORM                 := $(shell uname)
NAME                     := $(shell basename $(CURDIR))
GOFMT_FILES              ?= $$(find ./ -name '*.go' | grep -v vendor | grep -v externalmodels)
GOTEST_DIRECTORIES       ?= $$(find ./internal/ -type f -iname "*_test.go" -exec dirname {} \; | uniq)
APP_NAME                 := "postgres-exporter"
DOCKER_TAG               := "local"

export GOPRIVATE=github.com/form3tech-oss
export GOFLAGS=-mod=vendor

DOCKER_IMG ?= form3tech/$(APP_NAME)

.PHONY: build
build:
	@echo "==> Building..."
	@go install ./...

.PHONY: test
test: install-pact
	@echo "==> Executing tests..."
	@echo ${GOTEST_DIRECTORIES} | xargs -n1 go test --timeout 30m -v -count 1

.PHONY: goimports
goimports: install-goimports
	goimports -w $(GOFMT_FILES)

.PHONY: install-goimports
install-goimports:
	@type goimports >/dev/null 2>&1 || (cd /tmp && go get golang.org/x/tools/cmd/goimports && cd -)

.PHONY: vendor
vendor:
	@go mod tidy && go mod vendor && go mod verify

.PHONY: docker-build
docker-build:
	@echo "==> Building docker image..."
	docker build --build-arg APPNAME=postgres_exporter -f Dockerfile -t $(DOCKER_IMG):$(DOCKER_TAG) .

.PHONY: docker-publish
docker-publish: docker-build
	@echo "==> Logging in to the docker registry..."
	echo "$(DOCKER_PASSWORD)" | docker login -u "$(DOCKER_USERNAME)" --password-stdin
	@echo "==> Pushing built image..."
	docker push $(DOCKER_IMG):$(DOCKER_TAG)
	docker tag $(DOCKER_IMG):$(DOCKER_TAG) $(DOCKER_IMG):latest