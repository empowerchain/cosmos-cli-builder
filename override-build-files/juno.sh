#!/bin/bash

mkdir -p build
make build -B
mv ./bin/junod ./build