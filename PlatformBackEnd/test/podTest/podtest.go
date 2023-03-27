package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"path/filepath"

	"strconv"

	"runtime"

	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/shirou/gopsutil/mem"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type PodData struct {
	Podname   string `json:"podname"`
	Container string `json:"container_name"`
	Memory    string `json:"memory"`
	Cpu       string `json:"cpu"`
	Gpu       string `json:"gpu"`
	Memlim    string `json:"memlim"`
	Cpulim    string `json:"cpulim"`
	Gpulim    string `json:"gpulim"`
	Imagename string `json:"imagename"`
	Mountname string `json:"mountname"`
	Mountpath string `json:"mountpath"`
	Nodename  string `json:"nodename"`
	Namespace string `json:"namespace"`
}

const gpuMetricName = "nvidia.com/gpu"

// 初始化Kubernetes客户端
func InitK8S() (*kubernetes.Clientset, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// 使用kubeconfig中的当前上下文,加载配置文件
	config, err_config := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err_config != nil {
		glog.Error("Failed to use load kubeconfig, the error is %s ", err_config)
		return nil, err_config
	}

	// 创建clientset
	clientset, err_client := kubernetes.NewForConfig(config)
	if err_client != nil {
		glog.Error("Failed to create clientset, the error is %s ", err_client)
		return nil, err_client
	}

	return clientset, nil
}

// 获取内存和GPU
func GetAvailableMemoryAndGPU() (uint64, int, map[int]uint64, error) {
	// 获取系统可用内存
	memInfo, _ := mem.VirtualMemory()
	memAva := memInfo.Available / 1024 / 1024 / 1024

	// 获取cpu核数
	cpuCore := runtime.NumCPU()

	// 获取GPU显存
	err_init := nvml.Init()
	if err_init != nil {
		glog.Error("Failed to init nvml to get futher info, the error is %s", err_init)
		return 0, 0, nil, err_init
	}
	defer nvml.Shutdown()

	m := make(map[int]uint64)
	// 获取当前机器上显卡数量，并获取每张显卡的数据
	deviceCount, err_gpu := nvml.GetDeviceCount()
	if err_gpu != nil {
		glog.Error("Failed to get all GPU num, the error is %s", err_gpu)
		return 0, 0, nil, err_gpu
	}
	for i := uint(0); i < deviceCount; i++ {
		device, err_device := nvml.NewDeviceLite(uint(i))
		if err_device != nil {
			glog.Error("Failed to get GPU device, the error is %s", err_device)
			return 0, 0, nil, err_device
		}

		deviceStatus, _ := device.Status()
		usedMem := deviceStatus.Utilization.Memory
		avaMem := *device.Memory - uint64(*usedMem)
		m[0] = avaMem

		glog.Info("GPU %s, the avaMem is %s", i, avaMem)
	}

	return memAva, cpuCore, m, nil
}

func CreatePod(c *gin.Context) {
	var pod PodData
	err_bind := c.ShouldBindJSON(&pod)
	if err_bind != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err_bind.Error()})
		glog.Error("Failed to parse data form request, the error is %s", err_bind)
		return
	}

	// 获取当前可用的Mem、CPU和PU
	var avaGPU uint64
	m := make(map[int]uint64)
	avaMem, avaCPU, m, _ := GetAvailableMemoryAndGPU()
	for i := range m {
		avaGPU += m[i]
	}

	// 比较用户请求数和可用数
	am_str := strconv.FormatUint(avaMem, 10)
	ac_str := strconv.FormatInt(int64(avaCPU), 10)
	ag_str := strconv.FormatUint(avaGPU, 10)
	if (pod.Memory > am_str) || (pod.Cpu > ac_str) || (pod.Gpu > ag_str) ||
		(pod.Memory > pod.Memlim) || (pod.Cpu > pod.Cpulim) || (pod.Gpu > pod.Gpulim) {
		err := errors.New("sources required are larger than the avaliable!")
		c.AbortWithError(http.StatusForbidden, err)
		glog.Error("Failed to alloc sources to create pod, because the free sources are limited!")
		return
	}

	// 解析内存、CPU和GPU至k8s模式
	memReq, err_mem := resource.ParseQuantity(pod.Memory)
	if err_mem != nil {
		glog.Error("Failed to parse mem, the error is %s", err_mem)
		c.AbortWithError(http.StatusBadRequest, err_mem)
		return
	}
	memLim, err_meml := resource.ParseQuantity(pod.Memlim)
	if err_meml != nil {
		glog.Error("Failed to parse mem, the error is %s", err_meml)
		c.AbortWithError(http.StatusBadRequest, err_meml)
		return
	}

	cpuReq, err_cpu := resource.ParseQuantity(pod.Cpu)
	if err_cpu != nil {
		glog.Error("Failed to parse cpu, the error is %s", err_cpu)
		c.AbortWithError(http.StatusBadRequest, err_cpu)
		return
	}
	cpuLim, err_cpul := resource.ParseQuantity(pod.Cpulim)
	if err_cpul != nil {
		glog.Error("Failed to parse mem, the error is %s", err_cpul)
		c.AbortWithError(http.StatusBadRequest, err_cpul)
		return
	}

	gpuReq, err_gpu := resource.ParseQuantity(pod.Gpu)
	if err_gpu != nil {
		glog.Error("Failed to parse gpu, the error is %s", err_gpu)
		c.AbortWithError(http.StatusBadRequest, err_gpu)
		return
	}
	gpuLim, err_gpul := resource.ParseQuantity(pod.Gpulim)
	if err_gpul != nil {
		glog.Error("Failed to parse mem, the error is %s", err_gpul)
		c.AbortWithError(http.StatusBadRequest, err_gpul)
		return
	}

	// pod的yaml
	newPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: pod.Podname,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    pod.Container,
					Image:   pod.Imagename,
					Command: []string{"/bin/bash", "-ce", "tail -f /dev/null"},
					Env: []corev1.EnvVar{
						{
							Name:  "NVIDIA_DRIVER_CAPABILITIES",
							Value: "all",
						},
					},
					ImagePullPolicy: corev1.PullIfNotPresent,
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceMemory:              memReq,
							corev1.ResourceCPU:                 cpuReq,
							corev1.ResourceName(gpuMetricName): gpuReq,
						},
						Limits: corev1.ResourceList{
							corev1.ResourceMemory:              memLim,
							corev1.ResourceCPU:                 cpuLim,
							corev1.ResourceName(gpuMetricName): gpuLim,
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      pod.Mountname,
							MountPath: pod.Mountpath,
						},
					},
				},
			},
			NodeName: pod.Nodename,
			Volumes: []corev1.Volume{
				{
					Name: pod.Mountname,
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: pod.Mountpath,
						},
					},
				},
			},
		},
	}

	// 创建pod
	clientset, err_k8s := InitK8S()
	if err_k8s != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err_k8s.Error()})
		glog.Error("Failed to start k8s %s", err_k8s)
		return
	}
	pod_container, err := clientset.CoreV1().Pods(pod.Namespace).Create(context.Background(), newPod, metav1.CreateOptions{})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		glog.Error("Failed to create pod %s", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code: ":    1,
		"message: ": fmt.Sprintf("Succeed to create pod, its name is %v", pod_container.GetObjectMeta().GetName()),
	})
}

func Core() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token,Authorization,Token")
		c.Header("Access-Control-Allow-Methods", "POST,GET,OPTIONS")
		c.Header("Access-Control-Expose-Headers", "Content-Length,Access-Control-Allow-Origin,Access-Control-Allow-Headers,Content-Type")
		c.Header("Access-Control-Allow-Credentials", "True")
		//放行索引options
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		//处理请求
		c.Next()
	}
}

func main() {
	flag.Parse()
	defer glog.Flush()

	router := gin.Default()
	router.Use(Core())

	router.POST("/pod", CreatePod)

	router.Run(":8080")

}
