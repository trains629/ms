#!/bin/env bash

docker run --rm -e FLEX_REDIS="git.trains629.com:6379" \
-e FLEX_NSQ="git.trains629.com:4146" \
-e Flex_POSTGRES='{dbname: biger,user: biger,password: hao123456789,sslmode: disable,host: git.trains629.com}' \
ms-init:0.0.1 /usr/local/bin/init -end-points="git.trains629.com:2379"