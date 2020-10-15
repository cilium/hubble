FROM docker.io/library/golang:1.14.10-alpine3.12 as builder
WORKDIR /go/src/github.com/cilium/hubble
RUN apk add --no-cache binutils git make \
 && go get -d github.com/google/gops \
 && cd /go/src/github.com/google/gops \
 && git checkout -b v0.3.6 v0.3.6 \
 && go install \
 && strip /go/bin/gops
COPY . .
RUN make clean && make hubble

FROM docker.io/library/alpine:3.12
RUN addgroup -S hubble \
 && apk add --no-cache bash curl jq
COPY --from=builder /go/src/github.com/cilium/hubble/hubble /usr/bin
# cilium-sysdump expects gops to be installed in /bin
COPY --from=builder /go/bin/gops /bin
CMD ["/usr/bin/hubble", "serve"]
