GO := go
INSTALL = $(QUIET)install
BINDIR ?= /usr/local/bin
IMAGE_REPOSITORY ?= quay.io/covalent/hubble
CONTAINER_ENGINE ?= docker
TARGET=hubble
GIT_BRANCH != which git >/dev/null 2>&1 && git rev-parse --abbrev-ref HEAD
GIT_HASH != which git >/dev/null 2>&1 && git rev-parse --short HEAD

all: hubble

hubble:
	$(GO) build -ldflags "-X 'github.com/cilium/hubble/pkg.GitBranch=${GIT_BRANCH}' -X 'github.com/cilium/hubble/pkg.GitHash=$(GIT_HASH)'" -o $(TARGET)

install:
	groupadd -f hubble
	$(INSTALL) -m 0755 -d $(DESTDIR)$(BINDIR)
	$(INSTALL) -m 0755 ./hubble $(DESTDIR)$(BINDIR)

clean:
	rm -f $(TARGET)

test:
	go test -timeout=30s -cover $$(go list ./...)

lint: check-fmt ineffassign
ifeq (, $(shell which golint))
	$(error "golint not installed; you can install it with `go get -u golang.org/x/lint/golint`")
endif
	golint -set_exit_status $$(go list ./...)

check-fmt:
	./contrib/scripts/check-fmt.sh

ineffassign:
ifeq (, $(shell which ineffassign))
	$(error "ineffassign not installed; you can install it with `go get -u github.com/gordonklaus/ineffassign`")
endif
	ineffassign .

image:
	$(CONTAINER_ENGINE) build -t $(IMAGE_REPOSITORY)$(if $(IMAGE_TAG),:$(IMAGE_TAG)) .

.PHONY: all clean check-fmt image ineffassign install lint test hubble
