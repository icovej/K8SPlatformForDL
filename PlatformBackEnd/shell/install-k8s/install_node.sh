#!/bin/bash
#将kubelet-bootstrap用户绑定到系统集群角色
IP_MASTER=$1
IP_NODE1=$2
IP_NODE2=$3

#创建bootstrap用户，node加入群集时使用
kubectl create clusterrolebinding kubelet-bootstrap \
--clusterrole=system:node-bootstrapper \
--user=kubelet-bootstrap
#创建system:anonymous 用户，通过kubectl exec进入pods时使用，不配置无法进入pod
#kubelet.conf文件中要对应加入如下信息
#authentication:
#  anonymous:
#    enabled: true
kubectl create clusterrolebinding system:anonymous \
--clusterrole=cluster-admin \
--user=system:anonymous
#创建kubconfig文件
bash ./kubeconfig.sh $IP_MASTER ./

ssh root@$IP_NODE1 "mkdir -p /opt/kubernetes/{cfg,bin,ssl}"
scp bootstrap.kubeconfig kube-proxy.kubeconfig root@$IP_NODE1:/opt/kubernetes/cfg
scp ./kubernetes/server/bin/{kubelet,kube-proxy} root@$IP_NODE1:/opt/kubernetes/bin
scp ./kubelet.sh ./proxy.sh root@$IP_NODE1:/root
ssh root@$IP_NODE1 "/bin/bash /root/kubelet.sh $IP_NODE1"
ssh root@$IP_NODE1 "/bin/bash /root/proxy.sh $IP_NODE1"

ssh root@$IP_NODE2 "mkdir -p /opt/kubernetes/{cfg,bin,ssl}"
scp ./kubernetes/server/bin/{kubelet,kube-proxy} root@$IP_NODE2:/opt/kubernetes/bin
scp bootstrap.kubeconfig kube-proxy.kubeconfig root@$IP_NODE2:/opt/kubernetes/cfg
scp ./kubelet.sh ./proxy.sh root@$IP_NODE2:/root
ssh root@$IP_NODE2 "/bin/bash kubelet.sh $IP_NODE2"
ssh root@$IP_NODE2 "/bin/bash proxy.sh $IP_NODE2"