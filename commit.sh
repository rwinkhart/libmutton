#!/bin/sh
gofmt -l -w -s ./core/*.go ./sync/*.go ./libmuttonserver.go
git commit -am "$1"
git push
