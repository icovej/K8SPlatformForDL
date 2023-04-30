#/bin/bash
IP_MASTER=$1
IP_NODE1=$2
IP_NODE2=$3


/opt/etcd/bin/etcdctl \
--ca-file=/opt/etcd/ssl/ca.pem \
--cert-file=/opt/etcd/ssl/server.pem \
--key-file=/opt/etcd/ssl/server-key.pem \
--endpoints="https://$IP_MASTER:2379,https://$IP_NODE1:2379,https://$IP_NODE2:2379" \
set /coreos.com/network/config '{ "Network": "172.17.0.0/16", "Backend": {"Type": "vxlan"}}'

mkdir flannel-v0.12
tar -zxf flannel-v0.12.0-linux-amd64.tar.gz -C flannel-v0.12 

ssh $IP_NODE1 "mkdir -p /opt/kubernetes/{bin,cfg,ssl}"
scp ./flannel-v0.12/{flanneld,mk-docker-opts.sh} $IP_NODE1:/opt/kubernetes/bin
scp ./flannel.sh $IP_NODE1:/root
ssh $IP_NODE1 "bash /root/flannel.sh 'https://$IP_MASTER:2379,https://$IP_NODE1:2379,https://$IP_NODE2:2379'"

ssh $IP_NODE2 "mkdir -p /opt/kubernetes/{bin,cfg,ssl}"
scp ./flannel-v0.12/{flanneld,mk-docker-opts.sh} $IP_NODE2:/opt/kubernetes/bin
scp ./flannel.sh $IP_NODE2:/root
ssh $IP_NODE2 "bash /root/flannel.sh 'https://$IP_MASTER:2379,https://$IP_NODE1:2379,https://$IP_NODE2:2379'"