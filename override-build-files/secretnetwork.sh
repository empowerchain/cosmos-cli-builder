#!/bin/bash

mkdir -p build
make build_cli
mv ./secretcli ./build