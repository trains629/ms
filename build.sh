#!/usr/bin/env bash

MAIN_PATH=`pwd`

echo $MAIN_PATH

if [[ ! -d ./bin ]]; then
  mkdir -p ./bin
fi

function build(){
  service=${1}
  prefix=${3-ms}
  version=${2-0.0.1}
  if [[ ! -d ./$service ]]; then
    echo "don't find service "$service
    cd $MAIN_PATH
    return
  fi
  cd ./$service
  CGO_ENABLED=0 go build -o ../$service -ldflags="-s -w"
  echo $prefix-$service:$version
  if [[ ! -f ./$service ]]; then
    echo "don't find service "$service
    cd $MAIN_PATH
    return
  fi
  sudo docker build -f ./Dockerfile -t $prefix-$service:$version .
  mv ./$service $MAIN_PATH/bin/
  cd $MAIN_PATH
}

build writer
build init

