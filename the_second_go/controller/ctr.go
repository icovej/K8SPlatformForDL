package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"the_second_go/tools"

	"github.com/docker/docker/api/types"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	gpuMetricName = "nvidia.com/gpu"
	// 基础dockerfile路径
	srcfilepath = ""
)

// 镜像数据的结构体
type ImageData struct {
	Dstpath       string   `json:"dstpath"`
	Osversion     string   `json:"osversion"`
	Pythonversion string   `json:"pythonversion"`
	Imagearray    []string `json:"Imagearray"`
	Imagename     string   `json:"Imagename"`
}

// Dir数据结构体
type DirData struct {
	Dir   string `json:"dir"`
	Depth string `json:"max-depth"`
}

// Pod数据结构体
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

// 模型训练数据
type ModelData struct {
	Loss     uint64 `json:"loss"`
	Accuracy uint64 `json:"accuracy"`
}

// 镜像制作
func CreateImage(c *gin.Context) {
	var image_data ImageData
	// 解析数据
	err_bind := c.ShouldBindJSON(&image_data)
	if err_bind != nil {
		c.JSON(http.StatusMovedPermanently, gin.H{
			"code: ":    1,
			"message: ": err_bind.Error(),
		})
		glog.Error("Failed to parse data from request to struct, the error is %s", err_bind)
		return
	}

	// 创建用户的dockerfile文件
	dstFilepath := image_data.Dstpath

	err_create := tools.CopyFile(srcfilepath, dstFilepath)
	if err_create != nil {
		c.JSON(http.StatusMovedPermanently, gin.H{
			"code: ":    1,
			"message: ": err_create.Error(),
		})
		glog.Error("Failed to create dockerfile, the error is %s", err_create)
		return
	}

	// 选择系统
	osVersion := image_data.Osversion
	cmd := "FROM " + osVersion + "\n"
	err_version := tools.WriteAtHead(dstFilepath, cmd)
	if err_version != nil {
		c.JSON(http.StatusMovedPermanently, gin.H{
			"code: ":    1,
			"message: ": err_version.Error(),
		})
		glog.Error("Failed to write osVersion to dockerfile, the error is %s", err_version)
		return
	}

	// 选择python
	pyVersion := image_data.Pythonversion
	err_py := tools.WriteAtTail(dstFilepath, pyVersion)
	if err_py != nil {
		c.JSON(http.StatusMovedPermanently, gin.H{
			"code: ":    1,
			"message: ": err_py.Error(),
		})
		glog.Error("Failed to write pyVersion to dockerfile, the error is %s", err_py)
		return
	}

	// 选择镜像，拉取后将镜像写入dockerfile
	imageArray := image_data.Imagearray
	dockerClient, _ := tools.InitDocker()
	defer dockerClient.Close()
	for i := range imageArray {
		reader, err := dockerClient.ImagePull(context.Background(), imageArray[i], types.ImagePullOptions{})
		if err != nil {
			c.JSON(http.StatusMovedPermanently, gin.H{
				"code: ":    1,
				"image: ":   imageArray[i],
				"message: ": err.Error(),
			})
			glog.Error("Failed to pull docker image %s, the error is %s", imageArray[i], err)
			return
		}

		// 写入dockerfile
		err_image := tools.WriteAtTail(dstFilepath, imageArray[i])
		if err_image != nil {
			c.JSON(http.StatusMovedPermanently, gin.H{
				"code: ":    1,
				"message: ": err_image.Error(),
			})
			glog.Error("Failed to write image to dockerfile, the error is %s", err_image)
			return
		}
		glog.Info("Succeed to pull docker image %s", imageArray[i])
		defer reader.Close()
	}

	// 调用exec执行dockerfile，创建用户自定义镜像
	imageName := image_data.Imagename
	cmd = "docker build -f "
	_, err_exec := tools.ExecCommand(cmd, dstFilepath, " -t ", imageName)
	if err_exec != nil {
		c.JSON(http.StatusMovedPermanently, gin.H{
			"code: ":    1,
			"message: ": err_exec.Error(),
		})
		glog.Error("Failed to exec docker build, the error is %s", err_exec)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Succeed to build image: %s", imageName),
	})
	return
}

// 获取所有目录的容量大小，为普通用户选择工作目录
func GetDirInfo(c *gin.Context) {
	var Dir DirData
	err_bind := c.ShouldBindJSON(&Dir)
	if err_bind != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		glog.Error("Invalid request payload")
		return
	}

	dir := Dir.Dir
	depth := Dir.Depth
	output, err_exec := tools.ExecCommand("du -h --max-depth=", depth, dir)
	if err_exec != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err_exec.Error()})
		glog.Error("Failed to get %s info, the error is %s", dir, err_exec)
		return
	}

	lines := strings.Split(string(output), "\n")
	result := make(map[string]string)
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		result[fields[1]] = fields[0]
	}
	c.JSON(http.StatusOK, result)
}

// 创建容器
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
	avaMem, avaCPU, m, _ := tools.GetAvailableMemoryAndGPU()
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
	clientset, err_k8s := tools.InitK8S()
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

// 获取模型训练数据，进行可视化
func GetModelingData(c *gin.Context) {

}
