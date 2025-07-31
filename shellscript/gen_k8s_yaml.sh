#!/usr/bin/env bash

SVC=$1
ENV=$2
if [ -z "$SVC" ]; then
  echo "Usage: build_k8s_yaml.sh <svc> <env>"
  exit 1
fi
if [ -z "$ENV" ]; then
  echo "Usage: build_k8s_yaml.sh <svc> <env>"
  exit 1
fi

< ./deploy/k8s_manifest/deployment.yaml sed s/var_svc/go-${SVC}/g | sed s/var_env/${ENV}/g
