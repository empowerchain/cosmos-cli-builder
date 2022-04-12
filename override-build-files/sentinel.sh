#!/bin/bash

git checkout Makefile
mkdir -p build
printf "\nbuild: mod-vendor\n\tgo build -o build/ -mod=readonly -tags="${BUILD_TAGS}" -ldflags="${LD_FLAGS}" ./cmd/sentinelhub\n" >> Makefile
make build
git checkout Makefile