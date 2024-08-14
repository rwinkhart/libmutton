#!/bin/sh
# This script generates portable libmuttonserver release binaries for the following platforms:
# - Linux (x86_64_v1)
# - Linux (aarch64)
# - Windows (x86_64_v1)
# - Windows (aarch64)

mkdir -p ./1output
cd ..
GOOS=linux CGO_ENABLED=0 GOAMD64=v1 go build -o ./packaging/1output/libmuttonserver-linux-x86_64_v1 -ldflags="-s -w" -trimpath ./libmuttonserver.go
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o ./packaging/1output/libmuttonserver-linux-aarch64 -ldflags="-s -w" -trimpath ./libmuttonserver.go
GOOS=windows CGO_ENABLED=0 GOAMD64=v1 go build -o ./packaging/1output/libmuttonserver-windows-x86_64_v1.exe -ldflags="-s -w" -trimpath ./libmuttonserver.go
GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -o ./packaging/1output/libmuttonserver-windows-aarch64.exe -ldflags="-s -w" -trimpath ./libmuttonserver.go
