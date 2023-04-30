#!/bin/bash



systemctl stop kube-controller-manager kube-apiserver kube-scheduler
systemctl disable kube-controller-manager kube-apiserver kube-scheduler
rm -rf /opt/kubernetes/*
rm -f /usr/lib/systemd/system/kube-controller-manager.service
rm -f /usr/lib/systemd/system/kube-apiserver.service
rm -f /usr/lib/systemd/system/kube-scheduler.service
systemctl daemon-reload
