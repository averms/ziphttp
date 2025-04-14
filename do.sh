#!/bin/sh
set -eu

args="-ldflags -s -trimpath -tags osusergo,netgo"

GOAMD64=v3 go build $args -o target/ .
# GOOS=windows go build $args -o target/ .
# GOOS=darwin GOARCH=arm64 $args -o target/ .
