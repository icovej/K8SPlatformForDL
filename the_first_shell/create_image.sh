#!/bin/bash

# 从前端获取镜像名
image_name=$1

# 拉取镜像
docker pull $image_name

# 打包镜像
docker save $image_name > $image_name.tar
