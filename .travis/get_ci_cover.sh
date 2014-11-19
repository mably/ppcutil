#!/bin/bash
set -evx
if [ "$TRAVIS_GO_VERSION" = "tip" ]; then
    go get -v golang.org/x/tools/cmd/cover
else
    go get -v code.google.com/p/go.tools/cmd/cover
fi