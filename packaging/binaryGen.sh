#!/bin/sh
# This script generates portable libmuttonserver release binaries for the following platforms:
# - Linux (x86_64_v2)
# - Linux (arm64v8.0)
# - Linux (arm64v8.7)
# - Windows (x86_64_v2)
# - Windows (arm64v8.7)

mkdir -p ./1output
cd ..
GOOS=linux CGO_ENABLED=0 GOAMD64=v2 go build -o ./packaging/1output/libmuttonserver-linux-x86_64_v2 -ldflags="-s -w" -trimpath ./libmuttonserver.go
GOOS=linux GOARCH=arm64 GOARM64=v8.0 CGO_ENABLED=0 go build -o ./packaging/1output/libmuttonserver-linux-arm64v8.0 -ldflags="-s -w" -trimpath ./libmuttonserver.go
GOOS=linux GOARCH=arm64 GOARM64=v8.7 CGO_ENABLED=0 go build -o ./packaging/1output/libmuttonserver-linux-arm64v8.7 -ldflags="-s -w" -trimpath ./libmuttonserver.go
GOOS=windows CGO_ENABLED=0 GOAMD64=v2 go build -o ./packaging/1output/libmuttonserver-windows-x86_64_v2.exe -ldflags="-s -w" -trimpath ./libmuttonserver.go
GOOS=windows GOARCH=arm64 GOARM64=v8.7 CGO_ENABLED=0 go build -o ./packaging/1output/libmuttonserver-windows-arm64v8.7.exe -ldflags="-s -w" -trimpath ./libmuttonserver.go
