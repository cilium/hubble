GO := go
INSTALL = $(QUIET)install
BINDIR ?= /usr/local/bin
IMAGE_REPOSITORY ?= quay.io/covalent/hubble
CONTAINER_ENGINE ?= docker

all: hubble

hubble:
	$(GO) build -o $@ $^

install:
	groupadd -f hubble
	$(INSTALL) -m 0755 -d $(DESTDIR)$(BINDIR)
	$(INSTALL) -m 0755 ./hubble $(DESTDIR)$(BINDIR)

clean:
	rm -f $(TARGET)

test:
	go test -timeout=30s -cover $$(go list ./...)

lint:
	golint -set_exit_status $$(go list ./...)

check-fmt:
	./contrib/scripts/check-fmt.sh

ineffassign:
	ineffassign .

image:
	$(CONTAINER_ENGINE) build -t $(IMAGE_REPOSITORY)$(if $(IMAGE_TAG),:$(IMAGE_TAG)) .

.PHONY: all clean check-fmt image ineffassign install lint test hubble
