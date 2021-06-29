GO := go
GO_BUILD = CGO_ENABLED=0 $(GO) build
INSTALL = $(QUIET)install
BINDIR ?= /usr/local/bin
TARGET=hubble
VERSION=$(shell cat VERSION)
GIT_BRANCH = $(shell which git >/dev/null 2>&1 && git rev-parse --abbrev-ref HEAD)
GIT_HASH = $(shell which git >/dev/null 2>&1 && git rev-parse --short HEAD)
GO_TAGS ?=
IMAGE_REPOSITORY ?= quay.io/cilium/hubble
IMAGE_TAG ?= $(if $(findstring -dev,$(VERSION)),latest,v$(VERSION))
CONTAINER_ENGINE ?= docker
RELEASE_UID ?= $(shell id -u)
RELEASE_GID ?= $(shell id -g)

TEST_TIMEOUT ?= 5s

GOLANGCILINT_WANT_VERSION = 1.41.1
GOLANGCILINT_VERSION = $(shell golangci-lint version 2>/dev/null)

all: hubble

hubble:
	$(GO_BUILD) $(if $(GO_TAGS),-tags $(GO_TAGS)) -ldflags "-w -s -X 'github.com/cilium/hubble/pkg.GitBranch=${GIT_BRANCH}' -X 'github.com/cilium/hubble/pkg.GitHash=$(GIT_HASH)' -X 'github.com/cilium/hubble/pkg.Version=${VERSION}'" -o $(TARGET)

release:
	docker run --env "RELEASE_UID=$(RELEASE_UID)" --env "RELEASE_GID=$(RELEASE_GID)" --rm --workdir /hubble --volume `pwd`:/hubble docker.io/library/golang:1.16.5-alpine3.13 \
		sh -c "apk add --no-cache make && make local-release"

local-release: clean
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
				ARCHS='386 amd64'; \
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
	done; \
	if [ $$(id -u) -eq 0 -a -n "$$RELEASE_UID" -a -n "$$RELEASE_GID" ]; then \
		chown -R "$$RELEASE_UID:$$RELEASE_GID" release; \
	fi

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

ifneq (,$(findstring $(GOLANGCILINT_WANT_VERSION),$(GOLANGCILINT_VERSION)))
check:
	golangci-lint run
else
check:
	docker run --rm -v `pwd`:/app -w /app docker.io/golangci/golangci-lint:v$(GOLANGCILINT_WANT_VERSION) golangci-lint run
endif

image:
	$(CONTAINER_ENGINE) build $(DOCKER_FLAGS) -t $(IMAGE_REPOSITORY)$(if $(IMAGE_TAG),:$(IMAGE_TAG)) .

.PHONY: all hubble release install clean test bench check image
