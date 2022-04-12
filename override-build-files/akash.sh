#!/bin/bash

mkdir -p build
make akash
mv ./.cache/bin/akash ./build