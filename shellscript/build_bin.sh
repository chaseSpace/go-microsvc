#!/usr/bin/env bash

# e.g. user barbase  crontask-xxx
SVC=$1
#ENV=$2
if [ -z "$SVC" ]; then
  echo "错误的参数!"
  echo "示例: sh build_bin.sh <SVC>"
  exit 1
fi

SVC=$(echo "$SVC" | cut -d '-' -f 1)
#ARG=$(echo "$SVC" | cut -d '-' -f 2)


#export GOOS=linux

# bug!!!
#golangci-lint run --timeout 1m

mkdir -p .bin/
CGO_ENABLED=0 GOARCH=amd64 go build -o .bin/$SVC -ldflags "-w -extldflags -static" service/$SVC/main.go
