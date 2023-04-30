#!/bin/bash

#step1 安装必要的系统工具
#step2 添加软件源信息
#step3 更新源 安装docke-ce
#step4 启动docker
#显示docker 版本

IP_NODE1=$1
IP_NODE2=$2

for ip in {$IP_NODE1,$IP_NODE2}
	do
ssh root@$ip \
	 "yum install -y yum-utils device-mapper-persistent-data lvm2;
	yum-config-manager --add-repo http://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo; 
	yum makecache fast; 
	yum -y install docker-ce; 
	systemctl restart docker; 
	systemctl enable docker; 
	docker version " ;

done


