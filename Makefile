GO := go
GO_BUILD = CGO_ENABLED=0 $(GO) build
INSTALL = $(QUIET)install
BINDIR ?= /usr/local/bin
TARGET=hubble
VERSION=$(shell cat VERSION)
# homebrew uses the github release's tarball of the source that does not contain the '.git' directory.
GIT_BRANCH = $(shell command -v git >/dev/null 2>&1 && git rev-parse --abbrev-ref HEAD 2> /dev/null)
GIT_HASH = $(shell command -v git >/dev/null 2>&1 && git rev-parse --short HEAD 2> /dev/null)
GO_TAGS ?=
IMAGE_REPOSITORY ?= quay.io/cilium/hubble
IMAGE_TAG ?= $(if $(findstring -dev,$(VERSION)),latest,v$(VERSION))
CONTAINER_ENGINE ?= docker
RELEASE_UID ?= $(shell id -u)
RELEASE_GID ?= $(shell id -g)

TEST_TIMEOUT ?= 5s

# renovate: datasource=docker depName=golangci/golangci-lint
GOLANGCILINT_WANT_VERSION = v1.53.2
GOLANGCILINT_IMAGE_SHA = sha256:2f64f422a09e620e33c3f5781e20a2c8086ff4eb7225d4be865e03fa9c750f8a
GOLANGCILINT_VERSION = $(shell golangci-lint version 2>/dev/null)

# renovate: datasource=docker depName=library/golang
GOLANG_IMAGE_VERSION = 1.20.5-alpine3.17
GOLANG_IMAGE_SHA = sha256:5b56b276431ae41eb150f1728f9e621fdd6720acce0b0ff8500aeb057cd2e57e

all: hubble

hubble:
	$(GO_BUILD) $(if $(GO_TAGS),-tags $(GO_TAGS)) -ldflags "-w -s -X 'github.com/cilium/hubble/pkg.GitBranch=${GIT_BRANCH}' -X 'github.com/cilium/hubble/pkg.GitHash=$(GIT_HASH)' -X 'github.com/cilium/hubble/pkg.Version=${VERSION}'" -o $(TARGET)

release:
	$(CONTAINER_ENGINE) run --rm --workdir /hubble --volume `pwd`:/hubble docker.io/library/golang:$(GOLANG_IMAGE_VERSION)@$(GOLANG_IMAGE_SHA) \
		sh -c "apk add --no-cache setpriv make git && \
			/usr/bin/setpriv --reuid=$(RELEASE_UID) --regid=$(RELEASE_GID) --clear-groups make GOCACHE=/tmp/gocache local-release"

local-release: clean
	set -o errexit; \
	for OS in darwin linux windows; do \
		EXT=; \
		ARCHS=; \
		case $$OS in \
			darwin) \
				ARCHS='amd64 arm64'; \
				;; \
			linux) \
				ARCHS='386 amd64 arm arm64'; \
				;; \
			windows) \
				ARCHS='386 amd64 arm64'; \
				EXT=".exe"; \
				;; \
		esac; \
		for ARCH in $$ARCHS; do \
			echo Building release binary for $$OS/$$ARCH...; \
			test -d release/$$OS/$$ARCH|| mkdir -p release/$$OS/$$ARCH; \
			env GOOS=$$OS GOARCH=$$ARCH $(GO_BUILD) $(if $(GO_TAGS),-tags $(GO_TAGS)) -ldflags "-w -s -X 'github.com/cilium/hubble/pkg.Version=${VERSION}'" -o release/$$OS/$$ARCH/$(TARGET)$$EXT; \
			tar -czf release/$(TARGET)-$$OS-$$ARCH.tar.gz -C release/$$OS/$$ARCH $(TARGET)$$EXT; \
			(cd release && sha256sum $(TARGET)-$$OS-$$ARCH.tar.gz > $(TARGET)-$$OS-$$ARCH.tar.gz.sha256sum); \
		done; \
		rm -r release/$$OS; \
	done;

install: hubble
	$(INSTALL) -m 0755 -d $(DESTDIR)$(BINDIR)
	$(INSTALL) -m 0755 ./hubble $(DESTDIR)$(BINDIR)

clean:
	rm -f $(TARGET)
	rm -rf ./release

test:
	$(GO) test -timeout=$(TEST_TIMEOUT) -race -cover $$($(GO) list ./...)

bench:
	$(GO) test -timeout=30s -bench=. $$($(GO) list ./...)

ifneq (,$(findstring $(GOLANGCILINT_WANT_VERSION:v%=%),$(GOLANGCILINT_VERSION)))
check:
	golangci-lint run
else
check:
	$(CONTAINER_ENGINE) run --rm -v `pwd`:/app -w /app docker.io/golangci/golangci-lint:$(GOLANGCILINT_WANT_VERSION)@$(GOLANGCILINT_IMAGE_SHA) golangci-lint run
endif

image:
	$(CONTAINER_ENGINE) build $(DOCKER_FLAGS) -t $(IMAGE_REPOSITORY)$(if $(IMAGE_TAG),:$(IMAGE_TAG)) .

.PHONY: all hubble release install clean test bench check image
