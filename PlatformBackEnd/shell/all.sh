#!/bin/bash

# 分割前端字符串
string_split(){
    local string="$1"
    local i=0
    local array=()

    # 使用read命令分割字符串
    read -ra array <<< "$string"
    # 返回分割后的字符串数组
    echo "${array[@]}"
}

# 拉取&打包镜像函数
pull_and_build_image() {
    local image_name=$1
    local tag="latest"
    local image_full_name="$registry/$image_name:$tag"
    echo "Pulling and building image $image_full_name..."
    docker pull $image_full_name
    docker tag $image_full_name $image_name
}

# 创建镜像函数
# shell怎么才能拿到前端的传来的数据？
ceate_image(){
    # 前端用户选择的镜像
    local input="$1"
    # 镜像数组
    local image_names=()
    # 定义镜像仓库地址
    local registry="docker.io"

    # 调用split_string函数分割字符串
    image_names=($(split_string "$input"))
    # 输出分割后的字符串数组
    echo "${result[@]}"

    # 循环遍历镜像名数组并调用函数
    for name in "${image_names[@]}"
    do
        pull_and_build_image "$name"
    done
}
