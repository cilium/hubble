GO := go
INSTALL = $(QUIET)install
BINDIR ?= /usr/local/bin
IMAGE_REPOSITORY ?= quay.io/cilium/hubble
CONTAINER_ENGINE ?= docker
TARGET=hubble
VERSION=0.7.0-dev
GIT_BRANCH = $(shell which git >/dev/null 2>&1 && git rev-parse --abbrev-ref HEAD)
GIT_HASH = $(shell which git >/dev/null 2>&1 && git rev-parse --short HEAD)
GO_TAGS ?=
RELEASE_UID ?= $(shell id -u)
RELEASE_GID ?= $(shell id -g)

TEST_TIMEOUT ?= 5s

all: hubble

hubble:
	$(GO) build $(if $(GO_TAGS),-tags $(GO_TAGS)) -ldflags "-w -s -X 'github.com/cilium/hubble/pkg.GitBranch=${GIT_BRANCH}' -X 'github.com/cilium/hubble/pkg.GitHash=$(GIT_HASH)' -X 'github.com/cilium/hubble/pkg.Version=${VERSION}'" -o $(TARGET)

release:
	docker run --env "RELEASE_UID=$(RELEASE_UID)" --env "RELEASE_GID=$(RELEASE_GID)" --rm --workdir /hubble --volume `pwd`:/hubble docker.io/library/golang:1.14.7-alpine3.12 \
		sh -c "apk add --no-cache make && make local-release"

local-release: clean
	for OS in darwin linux windows; do \
		EXT=; \
		ARCHS=; \
		case $$OS in \
			darwin) \
				ARCHS='amd64'; \
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
			env GOOS=$$OS GOARCH=$$ARCH $(GO) build $(if $(GO_TAGS),-tags $(GO_TAGS)) -ldflags "-w -s -X 'github.com/cilium/hubble/pkg.Version=${VERSION}'" -o release/$$OS/$$ARCH/$(TARGET)$$EXT; \
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
	go test -timeout=$(TEST_TIMEOUT) -race -cover $$(go list ./...)

bench:
	go test -timeout=30s -bench=. $$(go list ./...)

check: check-fmt ineffassign lint staticcheck vet

check-fmt:
	./contrib/scripts/check-fmt.sh

ineffassign:
ifeq (, $(shell which ineffassign))
	$(error "ineffassign not installed; you can install it with `go get -u github.com/gordonklaus/ineffassign`")
endif
	ineffassign .

lint:
ifeq (, $(shell which golint))
	$(error "golint not installed; you can install it with `go get -u golang.org/x/lint/golint`")
endif
	golint -set_exit_status $$(go list ./...)

# Ignored staticcheck warnings:
# - SA1019 deprecation warnings: https://staticcheck.io/docs/checks#SA1019
# - ST1000 missing package comment: https://staticcheck.io/docs/checks#ST1000
staticcheck:
ifeq (, $(shell which staticcheck))
	$(error "staticcheck not installed; you can install it with `go get -u honnef.co/go/tools/cmd/staticcheck`")
endif
	staticcheck -checks="all,-SA1019,-ST1000" ./...

vet:
	go vet $$(go list ./...)

image:
	$(CONTAINER_ENGINE) build -t $(IMAGE_REPOSITORY)$(if $(IMAGE_TAG),:$(IMAGE_TAG)) .

.PHONY: all hubble release install clean test bench check check-fmt ineffassign lint vet image
