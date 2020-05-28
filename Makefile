GO := go
INSTALL = $(QUIET)install
BINDIR ?= /usr/local/bin
IMAGE_REPOSITORY ?= quay.io/covalent/hubble
CONTAINER_ENGINE ?= docker
TARGET=hubble
GIT_BRANCH != which git >/dev/null 2>&1 && git rev-parse --abbrev-ref HEAD
GIT_HASH != which git >/dev/null 2>&1 && git rev-parse --short HEAD
GO_TAGS ?=

TEST_TIMEOUT ?= 5s

all: hubble

hubble:
	$(GO) build $(if $(GO_TAGS),-tags $(GO_TAGS)) -ldflags "-w -s -X 'github.com/cilium/hubble/pkg.GitBranch=${GIT_BRANCH}' -X 'github.com/cilium/hubble/pkg.GitHash=$(GIT_HASH)'" -o $(TARGET)

install:
	$(INSTALL) -m 0755 -d $(DESTDIR)$(BINDIR)
	$(INSTALL) -m 0755 ./hubble $(DESTDIR)$(BINDIR)

clean:
	rm -f $(TARGET)

test:
	go test -timeout=$(TEST_TIMEOUT) -race -cover $$(go list ./...)

bench:
	go test -timeout=30s -bench=. $$(go list ./...)

check: check-fmt ineffassign lint vet

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

vet:
	go vet $$(go list ./...)

image:
	$(CONTAINER_ENGINE) build -t $(IMAGE_REPOSITORY)$(if $(IMAGE_TAG),:$(IMAGE_TAG)) .

.PHONY: all hubble install clean test bench check check-fmt ineffassign lint vet image
