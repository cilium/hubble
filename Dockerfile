FROM docker.io/library/golang:1.23.3-alpine3.19@sha256:36cc30986d1f9bc46670526fe6553b078097e562e196344dea6a075e434f8341 as builder
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
FROM docker.io/library/alpine:3.20.3@sha256:beefdbd8a1da6d2915566fde36db9db0b524eb737fc57cd1367effd16dc0d06d
RUN apk add --no-cache bash curl jq
COPY --from=builder /go/src/github.com/cilium/hubble/hubble /usr/bin
CMD ["/usr/bin/hubble"]
