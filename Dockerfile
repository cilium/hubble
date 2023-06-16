FROM docker.io/library/golang:1.20.5-alpine3.18@sha256:fd9d9d7194ec40a9a6ae89fcaef3e47c47de7746dd5848ab5343695dbbd09f8c as builder
WORKDIR /go/src/github.com/cilium/hubble
RUN apk add --no-cache git make
COPY . .
RUN make clean && make hubble

# NOTE: As of 2021-07-14, Alpine 3.11, 3.13 and 3.14 suffer from a bug in
# busybox[0] that affects busybox's nslookup implementation. Under certain
# conditions that typically depend on `/etc/resolv.conf` configuration,
# nslookup returns with exit code 1 instead of 0 even when the given name is
# resolved successfully. More information about the bug can be found on this
# thread[1].
# [0]: https://bugs.busybox.net/show_bug.cgi?id=12541
# [1]: https://github.com/gliderlabs/docker-alpine/issues/539
FROM docker.io/library/alpine:3.18.2@sha256:82d1e9d7ed48a7523bdebc18cf6290bdb97b82302a8a9c27d4fe885949ea94d1
RUN apk add --no-cache bash curl jq
COPY --from=builder /go/src/github.com/cilium/hubble/hubble /usr/bin
CMD ["/usr/bin/hubble"]
