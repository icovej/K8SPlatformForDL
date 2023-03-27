# Build Kubernetes-GPU Cluster

## First---Disable swap memory

This is what Kubernetes asks us to do, especially if you want to use its newer version.

```shell
sudo vi /etc/fstab
# comment which line has type of swap or auto
# for example
# /dev/fd0        /media/floppy0  auto    rw,user,noauto,exec,utf8 0       0
# after finishing modify it
reboot
```

## Second---Install docker

You'd better follow the [official method](https://docs.docker.com/desktop/install/linux-install/) to install docker.

But this kind of docker can just support CPU, so if you want use containers to run deep-learning model, you need to install docker of GPU. Nvidia gives this [method](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/install-guide.html#docker)

I have used above method successfully, BUT there is another way. According to some issues, after docker version 19, the Runtime of GPU docker has been integrated into docker and controlled by the -gpus parameter when used. But I don't know if it works for Kubernetes and I also give what I researched.

````shell
# Install the specified version of the dependency package
sudo apt-get install libnvidia-container1=1.4.0-1 libnvidia-container-tools=1.4.0-1 nvidia-container-toolkit=1.5.1-1

# Install Docker-CE using the official Docker script
curl https://get.docker.com | sh \
  && sudo systemctl --now enable docker

# Add stable repository and GPG key
distribution=$(. /etc/os-release;echo $ID$VERSION_ID) \
   && curl -s -L https://nvidia.github.io/nvidia-docker/gpgkey | sudo apt-key add - \
   && curl -s -L https://nvidia.github.io/nvidia-docker/$distribution/nvidia-docker.list | sudo tee /etc/apt/sources.list.d/nvidia-docker.list

# Update source
sudo apt-get update

# nvidia-docker2
sudo apt-get install -y nvidia-docker2

# Restart docker service
sudo systemctl restart docker

# Check if docker is active
sudo systemctl status docker

# Modify /etc/docker/daemon.json to enable the nvidia runtime as our default runtime on our node

sudo vim /etc/docker/daemon.json
```
{
    "default-runtime": "nvidia",
    "runtimes": {
        "nvidia": {
            "path": "/usr/bin/nvidia-container-runtime",
            "runtimeArgs": []
        }
    }
}
```

# Restart again
sudo systemctl restart docker

# Check if we can use args "--gpus"
docker run --help | grep -i gpus

# Run a container based on CUDA
sudo docker run --rm --gpus all nvidia/cuda:10.0-cudnn7-runtime-ubuntu16.04 nvidia-smi

````

## Third---Configur docker

This step meanly to modify docker's process isolation tool. Docker's tool is "cgroup", but Kubernetes's tool is "systemd". So this must be changed.

````shell
sudo vi /etc/docker/daemon.json
```
  {
    "exec-opts": [ "native.cgroupdriver=systemd" ]
  }
```
sudo systemctl daemon-reload
sudo systemctl restart docker
sudo systemctl status docker
````

Please make sure you have this file. If you don't find it, you need to uninstall docker completely and install it again.

If you are in China, you'd better change the docker download source. It's also in this file.

## Forth---Install Kubernetes

### Install GO

Our platform is based on GO, and in fact, in future, if you want to use or develop any CSI(Container Storage Interface), GO is essential. (I must say the method of Installing GO is the easiest one)

````shell
# "xxxx" means go's version you want, Please check [here](https://go.dev/dl/)
wget -c https://dl.google.com/go/goxxxxx.linux-amd64.tar.gz -O - | sudo tar -xz -C /usr/local
# 
~/ .profile
```
	export PATH=$PATH:/usr/local/go/bin
```
source ~/.profile
go version
````

### Install Kubernetes

1. Install tools of https and the three basic tools of Kubernetes

   ```shell
   sudo apt-get update && apt-get install -y apt-transport-https curl
   sudo apt-get install -y kubelet kubeadm kubectl --allow-unauthenticated
   # check if they are right
   kubeadm version
   ```

   After checking, if you get errors, like:

   ```shell
   No apt package "kubeadm", but there is a snap with that name.
   Try "snap install kubeadm"
   
   No apt package "kubectl", but there is a snap with that name.
   Try "snap install kubectl"
   
   No apt package "kubelet", but there is a snap with that name.
   Try "snap install kubelet"
   ```

   Open ` /etc/apt/sources.list ` , and add one line about the resource, like:

   `deb https://mirrors.aliyun.com/kubernetes/apt kubernetes-xenial main`

   And then, install them again.

   If there is an error like: 

   `The following signatures couldn't be verified because the public key is not available: NO_PUBKEY xxxxxxxx`

   Please run the following cmd:

   ```shell
   sudo apt-key adv --keyserver kerserver.ubuntu.com --recv-keys XXXXXXX
   ```

2. Init master node

   Before run `kubeadm init`, you can `docker pull` images what Kubernetes needs. Use following cmd to get image name and version. After pulling them, you can use `docker tag` to change their name to what Kubernetes asks

   ```shell
   kubeadm config images list
   ```

   Choose one machine as your master node (just which you like), and the run the following cmd:

   ```shell
   kubeadm init \
   --apiserver-advertise-address= \
   --image-repository registry.aliyuncs.com/google_containers \
   --pod-network-cidr= \
   --ignore-preflight-errors=
   ```

   The meanings of these parameters:

   - **--apiserver-advertise-addres:** Deployed address of apiserver, the main service of Kubernetes. It must be your master's IP.
   - **--image-repository:** Docker image source. If you are in China, you'd better use it.
   - **--pod-network-cidr:** This is the node network used by Kubernetes, because I use flannel as the Kubernetes network, so  fill in 10.244.0.0/16
   - **--ignore-preflight-errors:** Ignore errors encountered during initialization. I usually use it because I don't want to see some errors like `The number of cpus is not enough 2 cores`, so I can use `--ignore-preflight-errors=CpuNum`

   After this command is executed, you'll get information as following:

   ```shell
   Your Kubernetes master has initialized successfully!
    
   To start using your cluster, you need to run the following as a regular user:
    
     mkdir -p $HOME/.kube
     sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
     sudo chown $(id -u):$(id -g) $HOME/.kube/config
    
   You should now deploy a pod network to the cluster.
   Run "kubectl apply -f [podnetwork].yaml" with one of the options listed at:
     https://kubernetes.io/docs/concepts/cluster-administration/addons/
    
   You can now join any number of machines by running the following on each node
   as root:
    
   kubeadm join xxxxxxxxxxxx
   ```

   There are two important information:

   - Commands that can make you use cluster normally：

     ```shell
     mkdir -p $HOME/.kube
     sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
     sudo chown $(id -u):$(id -g) $HOME/.kube/config
     ```

   - Command, the last line, that runs on other machines so that you can add these machines to your cluster

     ```shell
     kubeadm join xxxxxxxxxxxx
     ```

     you don't need to save it forever. In future, if you forget it, run following cmd to get new **TOKEN** and **ca-cert-hash**

     ```shell
     kubeadm token create
     openssl x509 -pubkey -in /etc/kubernetes/pki/ca.crt | openssl rsa -pubin -outform der 2>/dev/null | openssl dgst -sha256 -hex | sed 's/^.* //'
     ```

     But new token can just work for 24 hours

3. Install Flannel

   ```shell
   kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/a70459be0084506e4ec919aa1c114638878db11b/Documentation/kube-flannel.yml
   ```

4. Check work

   Run the following cmd to check:

   ```shell
   # check status of cluster
   kubectl get cs
   # check nodes have been added to the cluster
   kubectl get nodes
   # if you find any node's status is NOTREADY, use cmd:
   kubectl describe node ${nodename}
   ```

   If you still have any errors, please google it, or you can give us an issue, we'll try our best to help you. In fact, I have use above methods to build at least three clusters, and everytime, I always meet new errors.

   I'll also give you a very easy yaml to create pod to check if this cluster can work succefully.

   ```yaml
   apiVersion: extensions/v1beta1
   kind: Deployment
   metadata:
     name: nginx-test
     namespace: kube-system
   spec:
     replicas: 1
     template:
       metadata:
         labels:
           k8s-app: nginx-test
       spec:
         containers:
         - name: nginx
           image: nginx
           imagePullPolicy: IfNotPresent
           ports:
           - containerPort: 80
             protocol: TCP
         nodeSelector:
           node-role.kubernetes.io/master: ""
         tolerations:
         - key: "node-role.kubernetes.io/master"
           effect: "NoSchedule"
   ---
   apiVersion: v1
   kind: Service
   metadata:
     name: proxy-nginx
     namespace: kube-system
   spec:
     type: NodePort
     ports:
     - port: 80
       targetPort: 80
       nodePort: 32767
     selector:
       k8s-app: proxy-nginx
   
   ```

   If its status is not `running`, you can use the following method to check it:

   ```shell
   kubectl -n ${namespace of you pod} describe pod ${podname} 
   journalctl -xeu kubelet
   ```

## Fifth---Set GPU

After build Kubernetes cluster, for supporting use GPU in Pod, we need to deploy Nvidia Device Plugin, which is develped by Nvidia. Just run one cmd:

```shell
$ kubectl create -f https://raw.githubusercontent.com/NVIDIA/k8s-device-plugin/v0.6.0/nvidia-device-plugin.yml
```

Sometimes, if you are in China, you will get this yaml very slowly, so I post all content of this yaml.

```yaml
# Copyright (c) 2019, NVIDIA CORPORATION.  All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: nvidia-device-plugin-daemonset
  namespace: kube-system
spec:
  selector:
    matchLabels:
      name: nvidia-device-plugin-ds
  updateStrategy:
    type: RollingUpdate
  template:
    metadata:
      # This annotation is deprecated. Kept here for backward compatibility
      # See https://kubernetes.io/docs/tasks/administer-cluster/guaranteed-scheduling-critical-addon-pods/
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ""
      labels:
        name: nvidia-device-plugin-ds
    spec:
      tolerations:
      # This toleration is deprecated. Kept here for backward compatibility
      # See https://kubernetes.io/docs/tasks/administer-cluster/guaranteed-scheduling-critical-addon-pods/
      - key: CriticalAddonsOnly
        operator: Exists
      - key: nvidia.com/gpu
        operator: Exists
        effect: NoSchedule
      # Mark this pod as a critical add-on; when enabled, the critical add-on
      # scheduler reserves resources for critical add-on pods so that they can
      # be rescheduled after a failure.
      # See https://kubernetes.io/docs/tasks/administer-cluster/guaranteed-scheduling-critical-addon-pods/
      priorityClassName: "system-node-critical"
      containers:
      - image: nvidia/k8s-device-plugin:1.11
        name: nvidia-device-plugin-ctr
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop: ["ALL"]
        volumeMounts:
          - name: device-plugin
            mountPath: /var/lib/kubelet/device-plugins
      volumes:
        - name: device-plugin
          hostPath:
            path: /var/lib/kubelet/device-plugins
```

And then, you need to edit `/etc/containerd/config.toml`, if `/etc/containerd` is not existed, please make it by hand flirstly.

```shell
mkdir /etc/containerd
```

Then, generate the default configuration:

```shell
containerd config default > /etc/containerd/config.toml
```

Then, modify two value:

```toml
[plugins."io.containerd.grpc.v1.cri".containerd.runtimes]

  [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
  	......
    runtime_type = "io.containerd.runtime.v1.linux" // This is the first

......

[plugins."io.containerd.runtime.v1.linux"]
  runtime = "nvidia-container-runtime" // This is the second
```

Lastly, restart containerd

```shell
sudo systemctl restart containerd
```

After its pod starts running, you can add `nvidia.com/gpu:` in your container's yaml

But this just support you to use GPU, instead of the specific graphics card. I tried to rewrite this plugin to support it, but I failed, and according to what I have researched, the management of hardware devices in Kubernetes can only handle the number of devices. This means if you want to use certain graphics card with strong computing power, Device Plugin cannot handle it. And I also don't find a good way to achieve it, maybe we can just use GPU-Sharing. It's really a pity.

