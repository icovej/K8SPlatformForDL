#!/bin/bash
IP_MASTER=$1
IP_NODE1=$2
IP_NODE2=$3


echo "-------untar k8s-install-pakage ---------"
tar -zxf kubernetes-server-linux-amd64.tar.gz

mkdir -p /opt/kubernetes/{bin,cfg,ssl}
 

echo "---------------cp k8s file -----------------"
cp ./kubernetes/server/bin/{kube-apiserver,kube-scheduler,kube-controller-manager} \
/opt/kubernetes/bin  

cp kubernetes/server/bin/kubectl /usr/bin

#create tonken  k8s-cert
echo "----------------create k8s-cert-------------------"
bash ./k8s-cert.sh $IP_MASTER $IP_NODE1 $IP_NODE2
cp ca.pem ca-key.pem server.pem server-key.pem /opt/kubernetes/ssl

#create_tonken
echo "----------------create tonken-------------------"
cat > /opt/kubernetes/cfg/token.csv <<EOF
0fb61c46f8991b718eb38d27b605b008,kubelet-bootstrap,10001,"system:kubelet-bootstrap"
EOF


#install api-server
echo "----------------install api-server----------------"
bash ./apiserver.sh  $IP_MASTER  https://$IP_MASTER:2379,https://$IP_NODE1:2379,https://$IP_NODE2:2379


#install controller-manager
echo "----------------install controller-manager---------"
bash ./controller-manager.sh $IP_MASTER

echo "----------------install scheduler------------------"
#install scheduler
bash ./scheduler.sh $IP_MASTER

