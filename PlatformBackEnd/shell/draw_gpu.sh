#!/bin/bash

# 容器名
container_name=$1

# 获取当前时间戳
current_timestamp=$(date +%s)

# 计算当前时间前 30 秒和此时间之后的时间戳
before_timestamp=$(($current_timestamp - 30))
after_timestamp=$(($current_timestamp + 30))

# 获取容器 GPU 使用情况，并将结果保存到变量中
gpu_usage=$(kubectl exec $container_name -- nvidia-smi --query-gpu=utilization.gpu --format=csv,noheader | awk '{print $1}')

# 获取容器 GPU 使用情况的时间戳，并将结果保存到变量中
gpu_timestamp=$(kubectl exec $container_name -- nvidia-smi --query-compute-apps=timestamp --format=csv,noheader | awk '{print $1}')

# 将时间戳转换为秒级时间戳
gpu_timestamp=$(date --date="$gpu_timestamp" +%s)

# 初始化数组
before_gpu_usage=()
after_gpu_usage=()

# 将 GPU 使用情况按照时间分别保存到数组中
for i in ${!gpu_timestamp[@]}; do
  if [ ${gpu_timestamp[$i]} -lt $before_timestamp ]; then
    before_gpu_usage+=(${gpu_usage[$i]})
  elif [ ${gpu_timestamp[$i]} -gt $after_timestamp ]; then
    after_gpu_usage+=(${gpu_usage[$i]})
  fi
done

# 将结果返回给前端
echo "Before: ${before_gpu_usage[@]}"
echo "After: ${after_gpu_usage[@]}"
