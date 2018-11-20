#!/bin/bash

function start_heplify_server_container()
{
    docker run --ulimit nofile=90000:90000 \
    --restart always  --name hepsrv --network="host" \
    --env-file $CONFIG_ENV_FILE \
    --log-driver json-file --log-opt max-size=10m --log-opt max-file=7 \
    -d registry.cn-beijing.aliyuncs.com/tinet-hub/heplify-server:$CONFIG_DOCKER_TAG
}

DEFAULT_ENV_FILE=/home/homer/config/env
DEFAULT_DOCKER_TAG=latest

START=$1
CONFIG_ENV_FILE=$2
CONFIG_DOCKER_TAG=$3

if [ "X$CONFIG_ENV_FILE" = "X" ];then
  if [ -f $DEFAULT_ENV_FILE ];then
    CONFIG_ENV_FILE=$DEFAULT_ENV_FILE
  else
    usage
    exit
  fi
fi

if [  "X$CONFIG_DOCKER_TAG" = "X" ];then
  CONFIG_DOCKER_TAG=$DEFAULT_DOCKER_TAG
fi

start_heplify_server_container