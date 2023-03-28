#!/bin/bash

# 输入的字符串
input_string="apple banana orange"

# 将字符串按照空格分割成数组
read -ra split_array <<< "$input_string"

# 输出分割后的数组
echo "Split array:"
for element in "${split_array[@]}"
do
    echo "$element"
done
