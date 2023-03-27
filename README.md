# Kubernetes(K8S) platform

（All commands are executed on Ubuntu）

## First----Build Kubernetes Cluster

Click [here](https://github.com/icovej/biyesheji/blob/master/platform_back_end/build_k8s_cluster.md)

## Second----Run platform

First, git clone this repository on your master node.

Then:

```shell
cd platform_back_end
sudo go run main.go srcfilepath="${your original dockerfile path}" -logdir="${glog path which you want to save glog}"
```





