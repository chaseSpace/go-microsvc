#!/usr/bin/env bash
set -e

# 此脚本支持在win、linux、mac环境下执行

# shellcheck disable=SC2188
<<EOF
使用示例：
sh build_pb.sh <svc_name>  # 生成指定服务的pb
sh build_pb.sh  # 生成所有pb
EOF

#if [ "$(ls -A ./proto)" ]; then
#   git submodule update --init
#fi

# e.g. user barbase  crontask-xxx
_svc=$1 # 指定生成的pb子目录，比如 user 对应 microsvc/protocol/svc/user

_pb_dir=$(echo "$_svc" | cut -d '-' -f 1)
_protoc_path=""

if [ ! -z $_pb_dir ]; then
  _pb_dir=${_pb_dir}pb
fi

if [[ $(uname) == "Linux" ]]; then
  echo "running on Linux"
  _protoc_path='./tool/linux/protoc_v24'
  chmod +x $_protoc_path/*
elif [[ $(uname) == "Darwin" ]]; then
  echo "running on macOS"
  _protoc_path='./tool/mac/protoc_v24'
  chmod +x $_protoc_path/*
elif [[ $(uname) == *MINGW* ]]; then
  echo "running on Windows"
  _protoc_path='./tool/win/protoc_v24'
else
  echo "unknown OS"
  exit 1
fi

PATH=$PATH:$_protoc_path

OUTPUT_DIR="./protocol"
mkdir -p $OUTPUT_DIR

FIRST_RUN_FILE='./protocol/.not_first_run'

case $_pb_dir in
clear)
  rm -rf ./protocol/*
  ;;
gatewaypb) # ignored
  exit
  ;;
*)
  # shellcheck disable=SC2016
  __protoc_gen='$_protoc_path/protoc -I ./proto/ -I ./proto/include/ \
          --go_out=$OUTPUT_DIR \
          --validate_out="lang=go,paths=source_relative:$OUTPUT_DIR" \
          --go_opt paths=source_relative \
          --go-grpc_out=$OUTPUT_DIR \
          --go-grpc_opt=paths=source_relative \
          --go-grpc_opt=require_unimplemented_servers=false'

  if [ ! -e $FIRST_RUN_FILE ]; then
    echo -n 'This is first build - '
    _pb_dir= # reset var, generate all
  fi

  if [[ -n $_pb_dir ]]; then
    echo "regenerate proto/svc/$_pb_dir/..." to $OUTPUT_DIR/svc/$_pb_dir

    rm -rf ./protocol/svc/$_pb_dir/*
    rm -rf ./protocol/svc/commonpb/*
    eval "$__protoc_gen proto/svc/commonpb/*.proto"
    eval "$__protoc_gen proto/svc/$_pb_dir/*.proto"

  else
    echo "regenerate [all] proto files..." to $OUTPUT_DIR/svc
    rm -rf ./protocol/*

    # 只能遍历生成（由于包含了go validator）
    for sub_dir in proto/svc/*; do
      eval "$__protoc_gen $sub_dir/*.proto"
    done

    # make a tag for first run
    if [ ! -e $FIRST_RUN_FILE ]; then
      echo ">> creating tag file: $FIRST_RUN_FILE"
      touch $FIRST_RUN_FILE
    fi
    #      --grpc-gateway_out=$OUTPUT_DIR \
    #      --grpc-gateway_opt logtostderr=true \
    #      --grpc-gateway_opt paths=source_relative \
    #      --grpc-gateway_opt generate_unbound_methods=true \
  fi

  # 去除 json tag 中的 omitempty
  if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i "" -e "s/,omitempty//g" ./protocol/svc/*/*.pb.go
  else
    sed -i -e "s/,omitempty//g" ./protocol/svc/*/*.pb.go
  fi
  ;;
esac

echo done.
