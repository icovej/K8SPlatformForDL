#/bin/bash

rm -rf /opt/kubernetes/cfg/{kubelet,kubelet.config,bootstrap.kubeconfig kube-proxy.kubeconfig }
rm -rf /opt/kubernetes/bin/{kubelet,kube-proxy} 
rm -rf /opt/kubernetes/ssl/*
systemctl stop kubelet.service
systemctl stop kube-proxy.service

rm -rf /usr/lib/systemd/system/kubelet.service
rm -rf /usr/lib/systemd/system/kubeproxy.service
systemctl daemon-reload