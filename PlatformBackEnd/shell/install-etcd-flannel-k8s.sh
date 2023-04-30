#!/bin/bash

IP_MASTER=192.168.31.71
IP_NODE1=192.168.31.72
IP_NODE2=192.168.31.73

echo "---STEP 1.-------------exec os-prepare-----------------------------\n"

/bin/bash ./os-prepare.sh

scp ./os-prepare.sh root@$IP_NODE1:/root
ssh  $IP_NODE1 "/bin/bash /root/os-prepare.sh"

scp ./os-prepare.sh root@$IP_NODE2:/root
ssh  $IP_NODE2 "/bin/bash /root/os-prepare.sh"

echo "---STEP 1.-------------exec os-prepare finished----------------------\n"



echo "---STEP 2.-------install etcd ----------------------------------------"

cd ./install-etcd
bash ./install_etcd.sh $IP_MASTER $IP_NODE1  $IP_NODE2

echo "---STEP 2.--------install etcd finished  ----------------------------\n\n"



echo "---STEP 3.-------install docker  ------------------------------------"

sleep 15
cd ../
bash install_docker.sh  $IP_NODE1  $IP_NODE2

echo "---STEP 3.------install docker finished ----------------------------\n\n"



echo "---STEP 4.--------install flannel  -----------------------------------"

sleep 15
cd ./install-flannel
bash install_flannel.sh $IP_MASTER $IP_NODE1  $IP_NODE2

echo "---STEP 4.--------install flannel finished  ----------------------------\n\n"



echo "---STEP 5.--------install k8s  -----------------------------------------"

cd ../install-k8s

echo "-----------STEP 5.1.--------install k8s master  -------------------------"
sleep 10
bash ./install_kube_master.sh $IP_MASTER $IP_NODE1  $IP_NODE2

echo "--------STEP 5.1.--------install k8s master finished  -------------------"
sleep 30

kubectl get cs
echo "--------STEP 5.2--------install k8s node  -------"

bash ./install_node.sh $IP_MASTER $IP_NODE1  $IP_NODE2

echo "--------STEP 5.2--------install k8s node finished   ---------------------"-

echo "------------get node status ---------------------------------------------"
kubectl get nodes