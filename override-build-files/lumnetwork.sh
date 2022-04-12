#!/bin/bash

git checkout Makefile
mkdir -p build
go mod tidy
printf "\ncustom-build: go.sum\n\t@go build -mod=readonly \$(BUILD_FLAGS) -o build/ ./cmd/lumd\n" >> Makefile
make custom-build
git checkout Makefile