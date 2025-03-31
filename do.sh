#!/bin/sh
set -eu

export GOAMD64=v3
go build -trimpath -tags osusergo,netgo -o target/ .
