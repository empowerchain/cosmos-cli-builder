#!/bin/bash

mkdir -p build
make build -B
mv ./bin/cerberusd ./build