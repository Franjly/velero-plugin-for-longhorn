PKG := github.com/ecatlabs/velero-plugin
BIN := velero-plugin

REGISTRY ?= ecatlabs
IMAGE    ?= $(REGISTRY)/velero-plugin
VERSION  ?= main 

GOOS   ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# local builds the binary using 'go build' in the local environment.
.PHONY: local
local: build-dirs
	CGO_ENABLED=0 go build -v -o _output/bin/$(GOOS)/$(GOARCH) .

# test runs unit tests using 'go test' in the local environment.
.PHONY: test
test:
	CGO_ENABLED=0 go test -v -timeout 60s ./...

# ci is a convenience target for CI builds.
.PHONY: ci
ci: verify-modules local test

# container builds a Docker image containing the binary.
.PHONY: container
container:
	docker build -t $(IMAGE):$(VERSION) .

# push pushes the Docker image to its registry.
.PHONY: push
push:
	@docker push $(IMAGE):$(VERSION)
ifeq ($(TAG_LATEST), true)
	docker tag $(IMAGE):$(VERSION) $(IMAGE):latest
	docker push $(IMAGE):latest
endif

# modules updates Go module files
.PHONY: modules
modules:
	go mod tidy

# verify-modules ensures Go module files are up to date
.PHONY: verify-modules
verify-modules: modules
	@if !(git diff --quiet HEAD -- go.sum go.mod); then \
		echo "go module files are out of date, please commit the changes to go.mod and go.sum"; exit 1; \
	fi

# build-dirs creates the necessary directories for a build in the local environment.
.PHONY: build-dirs
build-dirs:
	@mkdir -p _output/bin/$(GOOS)/$(GOARCH)

# clean removes build artifacts from the local environment.
.PHONY: clean
clean:
	@echo "cleaning"
	rm -rf _output
