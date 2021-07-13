FROM docker.io/library/golang:1.16.6-alpine3.14@sha256:a8df40ad1380687038af912378f91cf26aeabb05046875df0bfedd38a79b5499 as builder
WORKDIR /go/src/github.com/cilium/hubble
RUN apk add --no-cache git make
COPY . .
RUN make clean && make hubble

FROM docker.io/library/alpine:3.14.0@sha256:1775bebec23e1f3ce486989bfc9ff3c4e951690df84aa9f926497d82f2ffca9d
RUN apk add --no-cache bash curl jq
COPY --from=builder /go/src/github.com/cilium/hubble/hubble /usr/bin
CMD ["/usr/bin/hubble"]
