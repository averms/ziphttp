#!/bin/sh
set -eu

args="-ldflags -s -trimpath -tags osusergo,netgo"

GOAMD64=v3 go build $args -o target/ .
# GOOS=windows GOARCH=amd64 GOAMD64=v3 go build $args -o target/ .
# GOOS=linux GOARCH=arm64 go build $args -o target/ .
# GOOS=darwin GOARCH=arm64 go build $args -o target/ .
