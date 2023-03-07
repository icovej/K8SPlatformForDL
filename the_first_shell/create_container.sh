#!/bin/bash

# 获取用户输入的容器名、内存大小和GPU挂载信息
container_name=$1
memory_limit=$2
gpu_mount=$3

# 使用kubectl命令创建容器
kubectl run $container_name --image=my_image --limits=memory=$memory_limit --overrides='
{
  "spec": {
    "containers": [
      {
        "name": "'"$container_name"'",
        "resources": {
          "limits": {
            "nvidia.com/gpu": '"$gpu_mount"'
          }
        }
      }
    ]
  }
}'
