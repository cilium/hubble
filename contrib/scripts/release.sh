#!/usr/bin/env bash

set -ex
set -o pipefail

docker run --rm --workdir /hubble --volume `pwd`:/hubble docker.io/library/golang:1.14.4-alpine3.12 \
  apk add --no-cache make && make release
