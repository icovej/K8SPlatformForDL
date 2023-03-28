#!/bin/bash

# 定义保存数据的文件路径
data_file=~/data.txt

# 循环执行，每隔5秒获取GPU使用情况并保存到文件中
while true
do
  # 获取所有K8S容器的名称
  container_names=$(kubectl get pods -o=jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}')

  # 循环遍历容器名称，获取GPU使用情况并保存到文件中
  for name in $container_names
  do
    # 获取当前时间和GPU使用情况
    timestamp=$(date +%s)
    gpu_usage=$(kubectl exec $name nvidia-smi --query-gpu=utilization.gpu --format=csv,noheader)

    # 将数据写入文件
    echo "$name,$timestamp,$gpu_usage" >> $data_file
  done

  # 等待5秒
  sleep 5
done
