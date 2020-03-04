#!/bin/env bash

FLEX_REDIS="git.trains629.com:6379" \
FLEX_NSQ="git.trains629.com:4146" \
Flex_POSTGRES='{dbname: biger,user: biger,password: hao123456789,sslmode: disable,host: git.trains629.com}' \
go run ./*.go -end-points="0.0.0.0:2379"