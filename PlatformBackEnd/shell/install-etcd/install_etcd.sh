#/bin/bash=

IP_MASTER=$1
IP_NODE1=$2
IP_NODE2=$3





echo "1.------------------exec install_cfssl------------------------------\n"
/bin/bash  ./install_cfssl.sh

echo "2.------------------exec etcd-cert------------------------------------\n"
/bin/bash  ./etcd-cert.sh $IP_MASTER $IP_NODE1 $IP_NODE2

/usr/bin/mkdir -p /opt/etcd/{cfg,bin,ssl}
echo "3.------------------copy cert to /opt/etcd/ssl -----------------------\n"
cp ./{ca.pem,server-key.pem,server.pem} /opt/etcd/ssl

echo "4.------------------untar etcd-v3.3.10-linux-amd64---------------------\n"
tar -zxf ./etcd-v3.3.10-linux-amd64.tar.gz

echo "5.------------------cp etcd, etcdctl to /opt/etcd/bin-------------------\n"
cp ./etcd-v3.3.10-linux-amd64/{etcd,etcdctl} /opt/etcd/bin


echo "6.------------------install and start etcd on master --------------------\n"
/bin/bash ./etcd.sh etcd01 $IP_MASTER etcd02=https://$IP_NODE1:2380,etcd03=https://$IP_NODE2:2380

echo "7.------------------install and start etcd on node1------------------------\n"

cp /opt/etcd/cfg/etcd ./etcd.node1
cp /opt/etcd/cfg/etcd ./etcd.node2
sed -i 's/ETCD_NAME=\"etcd01\"/ETCD_NAME=\"etcd02\"/'  etcd.node1
sed -i "s/$IP_MASTER:2380\"/$IP_NODE1:2380\"/" etcd.node1
sed -i "s/$IP_MASTER:2379\"/$IP_NODE1:2379\"/" etcd.node1 

ssh  $IP_NODE1 "mkdir -p /opt/etcd/{cfg,bin,ssl}"

scp /opt/etcd/bin/{etcd,etcdctl} root@$IP_NODE1:/opt/etcd/bin
scp /opt/etcd/ssl/* root@$IP_NODE1:/opt/etcd/ssl
scp etcd.node1 root@$IP_NODE1:/opt/etcd/cfg/etcd
scp /usr/lib/systemd/system/etcd.service root@$IP_NODE1:/usr/lib/systemd/system/

ssh  $IP_NODE1 "systemctl daemon-reload"
ssh  $IP_NODE1 "systemctl enable etcd"
ssh  $IP_NODE1 "systemctl restart etcd" &

echo "8.------------------install and start etcd on node2------------------------\n"

sed -i 's/ETCD_NAME=\"etcd01\"/ETCD_NAME=\"etcd03\"/p' etcd.node2
sed -i "s/$IP_MASTER:2380\"/$IP_NODE2:2380\"/" etcd.node2
sed -i "s/$IP_MASTER:2379\"/$IP_NODE2:2379\"/" etcd.node2 


ssh  $IP_NODE2 "mkdir -p /opt/etcd/{cfg,bin,ssl}"
scp /opt/etcd/bin/{etcd,etcdctl} root@$IP_NODE2:/opt/etcd/bin
scp /opt/etcd/ssl/* root@$IP_NODE2:/opt/etcd/ssl
scp etcd.node2 root@$IP_NODE2:/opt/etcd/cfg/etcd
scp /usr/lib/systemd/system/etcd.service root@$IP_NODE2:/usr/lib/systemd/system/

ssh  $IP_NODE2 "systemctl daemon-reload" 
ssh  $IP_NODE2 "systemctl enable etcd"
ssh  $IP_NODE2 "systemctl restart etcd" &

echo "9.------------------this cluster status ------------------------\n"

sleep 15
/opt/etcd/bin/etcdctl \
--ca-file=ca.pem --cert-file=server.pem --key-file=server-key.pem \
--endpoints="https://$IP_MASTER:2379,https://$IP_NODE1:2379,https://$IP_NODE2:2379" \
cluster-health