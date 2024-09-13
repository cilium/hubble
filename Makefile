GO := go
GO_BUILD_FLAGS =
GO_TEST_FLAGS =
GO_BUILD = CGO_ENABLED=0 $(GO) build $(GO_BUILD_FLAGS)
GO_TEST = $(GO) test $(GO_TEST_FLAGS) -timeout=$(TEST_TIMEOUT)
INSTALL = $(QUIET)install
BINDIR ?= /usr/local/bin
SUBDIRS_HUBBLE_CLI := .
TARGET=hubble
VERSION=$(shell go list -f {{.Version}} -m github.com/cilium/cilium)
# homebrew uses the github release's tarball of the source that does not contain the '.git' directory.
GIT_BRANCH = $(shell command -v git >/dev/null 2>&1 && git rev-parse --abbrev-ref HEAD 2> /dev/null)
GIT_HASH = $(shell command -v git >/dev/null 2>&1 && git rev-parse --short HEAD 2> /dev/null)
GO_TAGS ?=
IMAGE_REPOSITORY ?= quay.io/cilium/hubble
IMAGE_TAG ?= $(if $(findstring -dev,$(VERSION)),latest,v$(VERSION))
CONTAINER_ENGINE ?= docker
RELEASE_UID ?= $(shell id -u)
RELEASE_GID ?= $(shell id -g)

RENOVATE_GITHUB_USER ?= renovate
RENOVATE_GITHUB_COM_TOKEN ?= $(shell gh auth token)

TEST_TIMEOUT ?= 5s

# renovate: datasource=docker depName=library/golang
GOLANG_IMAGE_VERSION = 1.23.1-alpine3.19
GOLANG_IMAGE_SHA = sha256:e0ea2a119ae0939a6d449ea18b2b1ba30b44986ec48dbb88f3a93371b4bf8750

# Add the ability to override variables
-include Makefile.override

all: hubble

hubble:
	$(MAKE) -C $(SUBDIRS_HUBBLE_CLI) hubble-bin

hubble-bin:
	$(GO_BUILD) $(if $(GO_TAGS),-tags $(GO_TAGS)) -ldflags "-w -s -X 'github.com/cilium/cilium/hubble/pkg.GitBranch=${GIT_BRANCH}' -X 'github.com/cilium/cilium/hubble/pkg.GitHash=$(GIT_HASH)' -X 'github.com/cilium/cilium/hubble/pkg.Version=${VERSION}'" -o $(TARGET) $(SUBDIRS_HUBBLE_CLI)

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
			env GOOS=$$OS GOARCH=$$ARCH $(GO_BUILD) $(if $(GO_TAGS),-tags $(GO_TAGS)) -ldflags "-w -s -X 'github.com/cilium/cilium/hubble/pkg.Version=${VERSION}'" -o release/$$OS/$$ARCH/$(TARGET)$$EXT; \
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
	$(GO_TEST) -race -cover $$($(GO) list ./...)

bench: TEST_TIMEOUT=30s
bench:
	$(GO_TEST) -bench=. $$($(GO) list ./...)

image:
	$(CONTAINER_ENGINE) build $(DOCKER_FLAGS) -t $(IMAGE_REPOSITORY)$(if $(IMAGE_TAG),:$(IMAGE_TAG)) .

renovate-local:
	@echo "Running renovate --platform=local"
	@docker run --rm -ti -e LOG_LEVEL=debug -e GITHUB_COM_TOKEN="$(RENOVATE_GITHUB_COM_TOKEN)" -v /tmp:/tmp -v $(PWD):/usr/src/app docker.io/renovate/renovate:full renovate --platform=local | tee renovate.log

.PHONY: all hubble release install clean test bench check image renovate-local
