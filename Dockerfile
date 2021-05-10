# Do not upgrade to alpine 3.13 as its nslookup tool returns 1, instead of 0
# for domain name lookups.
FROM docker.io/library/golang:1.16.4-alpine3.12@sha256:371dc6bf7e0c7ce112a29341b000c40d840aef1dbb4fdcb3ae5c0597e28f3061 as builder
WORKDIR /go/src/github.com/cilium/hubble
RUN apk add --no-cache git make
COPY . .
RUN make clean && make hubble

# Do not upgrade to alpine 3.13 as its nslookup tool returns 1, instead of 0
# for domain name lookups.
FROM docker.io/library/alpine:3.12.7@sha256:36553b10a4947067b9fbb7d532951066293a68eae893beba1d9235f7d11a20ad
RUN apk add --no-cache bash curl jq
COPY --from=builder /go/src/github.com/cilium/hubble/hubble /usr/bin
CMD ["/usr/bin/hubble"]
