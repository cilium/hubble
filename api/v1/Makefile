# Hijack PATH so we can run the repository's protoc-gen-go version
PROTOC := PATH=.:${PATH} protoc

HUBBLE_PROTO_SOURCES := \
	./flow/flow.proto \
	./observer/observer.proto

HUBBLE_GO_TARGETS := $(HUBBLE_PROTO_SOURCES:.proto=.pb.go) $(HUBBLE_PROTO_SOURCES:.proto=.pb.validate.go)

HUBBLE_PROTO_PATH := .

PROTO_DEPS = \
	-I external

RAW_GO_MAPPINGS := google/protobuf/any.proto=github.com/gogo/protobuf/types/
RAW_GO_MAPPINGS += google/protobuf/duration.proto=github.com/gogo/protobuf/types/
RAW_GO_MAPPINGS += google/protobuf/struct.proto=github.com/gogo/protobuf/types/
RAW_GO_MAPPINGS += google/protobuf/timestamp.proto=github.com/gogo/protobuf/types/
RAW_GO_MAPPINGS += google/protobuf/wrappers.proto=github.com/gogo/protobuf/types/

RAW_GO_MAPPINGS += flow/flow.proto=github.com/cilium/hubble/api/v1/flow/
RAW_GO_MAPPINGS += observer/observer.proto=github.com/cilium/hubble/api/v1/observer/

# Add mapping separators and remove the trailing slash
# but first create "/ " and ",M"
file_sep := /
file_sep +=
map_sep := ,M
GO_MAPPINGS := $(patsubst %/,%,$(map_sep)$(subst $(file_sep),$(map_sep),$(RAW_GO_MAPPINGS)))

$(HUBBLE_GO_TARGETS): $(HUBBLE_PROTO_SOURCES) Makefile deps
	$(QUIET)set -e; \
	for proto in $(HUBBLE_PROTO_SOURCES) ; do \
		$(PROTOC) -I $(HUBBLE_PROTO_PATH) $(PROTO_DEPS) \
			--go_out=plugins=grpc$(GO_MAPPINGS):. --validate_out=lang=go$(GO_MAPPINGS):. --go-json_out=orig_name:. \
			$${proto}; \
	done

.PHONY: all
all: $(HUBBLE_GO_TARGETS)

.PHONY: clean
clean:
	rm -f $(HUBBLE_GO_TARGETS)

.PHONY: deps
deps:
	@$(PROTOC) --version | grep 3.9.1 || (echo "Error: protoc 3.9.1 is required" && exit 1)
