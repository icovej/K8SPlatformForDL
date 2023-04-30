#!/bin/bash



NODE1_IP=192.168.31.72
NODE2_IP=192.168.31.73

echo "------detele master ----------"
rm -rf /opt/etcd
systemctl stop etcd.service
rm -f /usr/lib/systemd/system/etcd.service
systemctl daemon-reload

echo "------detele node1----------"
ssh  $NODE1_IP "rm -rf /opt/etcd"
ssh  $NODE1_IP "systemctl stop etcd"
ssh  $NODE1_IP "rm -f /usr/lib/systemd/system/etcd.service"
ssh  $NODE1_IP "systemctl daemon-reload"

echo "------detele node2----------"
ssh  $NODE2_IP "rm -rf /opt/etcd"
ssh  $NODE2_IP "systemctl stop etcd"
ssh  $NODE2_IP "rm -f /usr/lib/systemd/system/etcd.service"
ssh  $NODE2_IP "systemctl daemon-reload"
