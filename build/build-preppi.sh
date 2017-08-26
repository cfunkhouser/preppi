#!/bin/bash
set -e

WORKDIR="/opt/gopath/src/github.com/cfunkhouser/preppi" ; cd "${WORKDIR}"

# Ensure all deps are available
dep ensure -update

# Build
env GOOS=linux GOARCH=arm GOARM=7 go build -v \
  -ldflags="-X main.buildID=$(git describe --always --long --dirty)" \
  -o build/out/bin/preppi-linux-armv7 main.go
