# Hijack PATH so we can run the repository's protoc-gen-go version
PROTOC := PATH=.:${PATH} protoc

HUBBLE_PROTO_SOURCES := \
	./observer.proto

HUBBLE_PROTO_DIRS := $(sort $(dir $(HUBBLE_PROTO_SOURCES)))

GO_OUT := observer

HUBBLE_GO_TARGETS := $(addprefix $(GO_OUT)/,$(HUBBLE_PROTO_SOURCES:.proto=.pb.go)) \
	$(addprefix $(GO_OUT)/,$(HUBBLE_PROTO_SOURCES:.proto=.pb.validate.go))

HUBBLE_PROTO_PATH := .

PROTO_DEPS = \
	-I ../../vendor \
	-I external

RAW_GO_MAPPINGS := google/protobuf/any.proto=github.com/gogo/protobuf/types/
RAW_GO_MAPPINGS += google/protobuf/duration.proto=github.com/gogo/protobuf/types/
RAW_GO_MAPPINGS += google/protobuf/struct.proto=github.com/gogo/protobuf/types/
RAW_GO_MAPPINGS += google/protobuf/timestamp.proto=github.com/gogo/protobuf/types/
RAW_GO_MAPPINGS += google/protobuf/wrappers.proto=github.com/gogo/protobuf/types/

# Add mapping separators and remove the trailing slash
# but first create "/ " and ",M"
file_sep := /
file_sep +=
map_sep := ,M
GO_MAPPINGS := $(patsubst %/,%,$(map_sep)$(subst $(file_sep),$(map_sep),$(RAW_GO_MAPPINGS)))

$(HUBBLE_GO_TARGETS): $(HUBBLE_PROTO_SOURCES) Makefile deps
	$(QUIET)set -e; \
	for path in $(HUBBLE_PROTO_DIRS) ; do \
		$(PROTOC) -I $(HUBBLE_PROTO_PATH) $(PROTO_DEPS) --go_out=plugins=grpc$(GO_MAPPINGS):$(GO_OUT) --validate_out=lang=go$(GO_MAPPINGS):$(GO_OUT) --go-json_out=orig_name:$(GO_OUT) $${path}*.proto; \
	done

.PHONY: all
all: $(HUBBLE_GO_TARGETS)

.PHONY: clean
clean:
	rm -fr $(GO_OUT)/*

.PHONY: deps
deps:
	@$(PROTOC) --version | grep 3.9.1 || (echo "Error: protoc 3.9.1 is required" && exit 1)
