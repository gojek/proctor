#!/usr/bin/env bash

mkdir -p $GOPATH/bin
go get -u github.com/golang/lint/golint
which glide > /dev/null 2>&1
if [ "0" -ne "$?" ]; then curl https://glide.sh/get | sh; fi
