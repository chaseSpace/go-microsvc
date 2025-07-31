#!/bin/bash
set -e

SVC=$1
MICRO_SVC_ENV=$2
TAG=$3

<<comment
脚本用法：sh build_image.sh SVC ENV [tag]
示例：
  sh build_image.sh user dev  # 开发/测试环境可不用指定tag，使用latest
  sh build_image.sh user dev 1.0.0  # 生产环境需要指定tag
comment

# 配置区
IMAGE_REPO=leigg # 替换为你的镜像仓库，默认使用官方docker仓库
# -----

if [ -z "$SVC" ] || [ -z "$MICRO_SVC_ENV" ]; then
  echo "错误的参数!"
  echo "示例: sh build_image.sh <SVC> <ENV> [tag]"
  exit 1
  fi

if [ ! -z "$TAG" ]; then
  TAG=":$TAG"
else
  TAG=1.0.0
fi

# windows执行会卡死
golangci-lint run --timeout 1m

_imageRemote=$IMAGE_REPO/go-$SVC$TAG

docker build --build-arg SVC=$SVC --build-arg MICRO_SVC_ENV=$MICRO_SVC_ENV . -t $_imageRemote
docker push $_imageRemote