#!/bin/sh
set -eu

go build -trimpath -tags osusergo,netgo -o target/ .
