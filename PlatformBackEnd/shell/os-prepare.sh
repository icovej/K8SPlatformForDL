#!/bin/bash
echo "-------------diable firewalld------------- "
systemctl stop firewalld
systemctl disable firewalld

echo "-------------disable selinux -------------"
sed -i 's/enforcing/disabled/' /etc/selinux/config 
setenforce 0

echo "-------------disable swap-------------"

swapoff -a

sed -i /swap/s/^/#/ /etc/fstab

echo "-------------change system config-------------"
cat > /etc/sysctl.d/k8s.conf << EOF
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF
sysctl --system