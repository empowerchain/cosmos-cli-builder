#!/bin/bash

git checkout Makefile
mkdir -p build
printf "\ncustom-build:\n\tgo build \${BUILD_FLAGS} -o build/persistenceCore ./node\n" >> Makefile
make custom-build
git checkout Makefile