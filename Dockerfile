FROM docker.io/library/golang:1.21.4-alpine3.18@sha256:70afe55365a265f0762257550bc38440e0d6d6b97020d3f8c85328f00200dd8e as builder
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
FROM docker.io/library/alpine:3.18.5@sha256:34871e7290500828b39e22294660bee86d966bc0017544e848dd9a255cdf59e0
RUN apk add --no-cache bash curl jq
COPY --from=builder /go/src/github.com/cilium/hubble/hubble /usr/bin
CMD ["/usr/bin/hubble"]
