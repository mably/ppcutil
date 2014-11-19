#!/bin/bash
set -evx
if [ "$TRAVIS_GO_VERSION" = "tip" ]; then
    go get -v golang.org/x/tools/cmd/vet
else
    go get -v code.google.com/p/go.tools/cmd/vet
fi