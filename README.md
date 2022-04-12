# Cosmos CLI builder

This repo holds binaries for many different cosmos chains.

It is essentially a Go script that fetches up-to-date information (such as current version tag, etcg) from the brilliant cosmos.directory API's and builds binaries from source for as many platforms and architectures as possible. This doesn't always work, so keep in mind that you might not find all builds for all chains here.

All binaries are built in Github Actions so you can verify that they are built from the correct source code.

The lastest binaires can be found here: https://github.com/empowerchain/cosmos-cli-builder/releases/tag/latest