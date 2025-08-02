#!/usr/bin/env bash
set -e

# e.g. user barbase  crontask-xxx
SVC=$1
K8S=$2
if [ -z "$SVC" ]; then
  echo "错误的参数!"
  echo "示例: sh build_bin.sh <SVC>"
  exit 1
fi

VAR_TAGS=""
if [ "$K8S" = 1 ]; then
  VAR_TAGS="-tags k8s"
fi

SVC=$(echo "$SVC" | cut -d '-' -f 1)
#ARG=$(echo "$SVC" | cut -d '-' -f 2)

#export GOOS=linux

# bug!!!
#golangci-lint run --timeout 1m

mkdir -p .bin/

CMD="CGO_ENABLED=0 GOARCH=amd64 go build $VAR_TAGS -o .bin/$SVC -ldflags '-w -extldflags -static' service/$SVC/main.go"

# 打印
echo "$CMD"

# 执行
eval "$CMD"