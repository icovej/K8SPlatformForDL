package controller

import (
	"context"
	"fmt"
	"net/http"
	"platform_back_end/data"
	"platform_back_end/tools"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Create Image
func CreateImage(c *gin.Context) {
	var image_data data.ImageData
	// Parse data that from front-end
	err_bind := c.ShouldBindJSON(&image_data)
	if err_bind != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code: ":    http.StatusBadRequest,
			"message: ": fmt.Sprintf("Invalid request payload, err is %v", err_bind.Error()),
		})
		glog.Error("Method CreateImage gets invalid request payload")
		return
	}

	// Create user's dockerfile
	dstFilepath := image_data.Dstpath

	err_create := tools.CopyFile(data.Srcfilepath, dstFilepath)
	if err_create != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code: ":    http.StatusMethodNotAllowed,
			"message: ": err_create.Error(),
		})
		glog.Error("Failed to create dockerfile, the error is %v", err_create)
		return
	}

	// Import OS used in user's pod
	osVersion := image_data.Osversion
	statement := "FROM " + osVersion + "\n"
	err_version := tools.WriteAtBeginning(dstFilepath, []byte(statement))
	if err_version != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code: ":    http.StatusMethodNotAllowed,
			"message: ": err_version.Error(),
		})
		glog.Error("Failed to write osVersion to dockerfile, the error is %v", err_version)
		return
	}

	// Import python used in user's pod
	pyVersion := image_data.Pythonversion
	err_py := tools.WriteAtTail(dstFilepath, pyVersion)
	if err_py != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code: ":    http.StatusMethodNotAllowed,
			"message: ": err_py.Error(),
		})
		glog.Error("Failed to write PyVersion to dockerfile, the error is %v", err_py)
		return
	}

	// Import images used in user's pod
	// And write into dockerfile whoes path is user's working path
	imageArray := image_data.Imagearray
	for i := range imageArray {
		err_image := tools.WriteAtTail(dstFilepath, imageArray[i])
		if err_image != nil {
			c.JSON(http.StatusMethodNotAllowed, gin.H{
				"code: ":    http.StatusMethodNotAllowed,
				"message: ": err_image.Error(),
			})
			glog.Error("Failed to write image to dockerfile, the error is %v", err_image)
			return
		}
	}

	// Create dockerfile
	imageName := image_data.Imagename
	cmd := "docker"
	_, err_exec := tools.ExecCommand(cmd, "build", "-t", imageName, "-f", dstFilepath, ".")
	if err_exec != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code: ":    http.StatusMethodNotAllowed,
			"message: ": err_exec.Error(),
		})
		glog.Error("Failed to exec docker build, the error is %v", err_exec)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code: ":    http.StatusOK,
		"message: ": fmt.Sprintf("Succeed to build image: %v", imageName),
	})
	return
}

// Get all dir infor which user request
func GetDirInfo(c *gin.Context) {
	var Dir data.DirData
	err_bind := c.ShouldBindJSON(&Dir)
	if err_bind != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code: ":    http.StatusBadRequest,
			"message: ": fmt.Sprintf("Invalid request payload, err is %v", err_bind.Error()),
		})
		glog.Error("Method GetDirInfo gets invalid request payload")
		return
	}

	dir := Dir.Dir
	depth := Dir.Depth
	output, err_exec := tools.ExecCommand("du -h --max-depth=", depth, dir)
	if err_exec != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code: ":    http.StatusMethodNotAllowed,
			"message: ": err_exec.Error(),
		})
		glog.Error("Failed to get %v info, the error is %v", dir, err_exec)
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

// Create Pod
func CreatePod(c *gin.Context) {
	var pod data.PodData
	err_bind := c.ShouldBindJSON(&pod)
	if err_bind != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code: ":    http.StatusBadRequest,
			"message: ": fmt.Sprintf("Invalid request payload, err is %v", err_bind.Error()),
		})
		glog.Error("Method CreatePod gets invalid request payload")
		return
	}

	// Get avaliable Mem, CPU and PU
	var avaGPU uint64
	m := make(map[int]uint64)
	avaMem, avaCPU, m, _ := tools.GetAvailableMemoryAndGPU()
	for i := range m {
		avaGPU += m[i]
	}

	// Compare the value user request and the avaliable
	// TODO:the logic of there has a little troubles, need to fix.
	// main problem is what we get from request has different unit, maybe need match that
	am_str := strconv.FormatUint(avaMem, 10)
	ac_str := strconv.FormatInt(int64(avaCPU), 10)
	ag_str := strconv.FormatUint(avaGPU, 10)
	if (pod.Memory > am_str) || (pod.Cpu > ac_str) || (pod.Gpu > ag_str) ||
		(pod.Memory > pod.Memlim) || (pod.Cpu > pod.Cpulim) || (pod.Gpu > pod.Gpulim) {
		err := errors.New("sources required are larger than the avaliable!")
		c.JSON(http.StatusBadRequest, gin.H{
			"code: ":    http.StatusBadRequest,
			"message: ": err.Error(),
		})
		glog.Error("Failed to alloc sources to create pod, because the free sources are limited!")
		return
	}

	// Parse mem、CPU and GPU to k8s mod
	memReq, err_mem := resource.ParseQuantity(pod.Memory)
	if err_mem != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code: ":    http.StatusBadRequest,
			"message: ": err_mem.Error(),
		})
		glog.Error("Failed to parse mem, the error is %v", err_mem)
		return
	}
	memLim, err_meml := resource.ParseQuantity(pod.Memlim)
	if err_meml != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code: ":    http.StatusBadRequest,
			"message: ": err_meml.Error(),
		})
		glog.Error("Failed to parse mem, the error is %v", err_meml)
		return
	}

	cpuReq, err_cpu := resource.ParseQuantity(pod.Cpu)
	if err_cpu != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code: ":    http.StatusBadRequest,
			"message: ": err_cpu.Error(),
		})
		glog.Error("Failed to parse cpu, the error is %v", err_cpu)
		return
	}
	cpuLim, err_cpul := resource.ParseQuantity(pod.Cpulim)
	if err_cpul != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code: ":    http.StatusBadRequest,
			"message: ": err_cpul.Error(),
		})
		glog.Error("Failed to parse mem, the error is %v", err_cpul)
		return
	}

	gpuReq, err_gpu := resource.ParseQuantity(pod.Gpu)
	if err_gpu != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code: ":    http.StatusBadRequest,
			"message: ": err_gpu.Error(),
		})
		glog.Error("Failed to parse gpu, the error is %v", err_gpu)
		return
	}
	gpuLim, err_gpul := resource.ParseQuantity(pod.Gpulim)
	if err_gpul != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code: ":    http.StatusBadRequest,
			"message: ": err_gpu.Error(),
		})
		glog.Error("Failed to parse mem, the error is %v", err_gpul)
		return
	}

	// form pod's yaml
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
							corev1.ResourceMemory:                   memReq,
							corev1.ResourceCPU:                      cpuReq,
							corev1.ResourceName(data.GpuMetricName): gpuReq,
						},
						Limits: corev1.ResourceList{
							corev1.ResourceMemory:                   memLim,
							corev1.ResourceCPU:                      cpuLim,
							corev1.ResourceName(data.GpuMetricName): gpuLim,
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

	// create pod
	clientset, err_k8s := tools.InitK8S()
	if err_k8s != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code: ":    http.StatusMethodNotAllowed,
			"message: ": err_k8s.Error(),
		})
		glog.Error("Failed to start k8s %v", err_k8s)
		return
	}
	pod_container, err := clientset.CoreV1().Pods(pod.Namespace).Create(context.Background(), newPod, metav1.CreateOptions{})
	if err != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code: ":    http.StatusMethodNotAllowed,
			"message: ": err.Error(),
		})
		glog.Error("Failed to create pod %v", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code: ":    http.StatusOK,
		"message: ": fmt.Sprintf("Succeed to create pod, its name is %v", pod_container.GetObjectMeta().GetName()),
	})
}
