package controller

import (
	"PlatformBackEnd/data"
	"PlatformBackEnd/tools"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Create Pod
// It's a little bulky, we'll fix it
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
	// the unit of avaMem and avaGPU is bytes, the unit of avaCPU is core
	avaMem, avaCPU, m, _ := tools.GetAvailableMemoryAndGPU()
	for i := range m {
		avaGPU += m[i]
	}

	// Compare the value user request and the avaliable
	ac_str := strconv.FormatInt(int64(avaCPU), 10)
	ag_str := strconv.FormatUint(avaGPU, 10)

	// according to user's request, transform user's value to Bytes
	memValue, memUnit := tools.GetLastTwoChars(pod.Memory)
	var pod_Memory int64
	if memUnit == "Gi" {
		tmp, _ := strconv.ParseFloat(memValue, 64)
		pod_Memory = tools.GiBToBytes(tmp)
	} else if memUnit == "Mi" {
		tmp, _ := strconv.ParseFloat(memValue, 64)
		pod_Memory = tools.MiBToBytes(tmp)
	}

	memlValue, memlUnit := tools.GetLastTwoChars(pod.Memlim)
	var pod_Lmemory int64
	if memlUnit == "Gi" {
		tmp, _ := strconv.ParseFloat(memlValue, 64)
		pod_Lmemory = tools.GiBToBytes(tmp)
	} else if memUnit == "Mi" {
		tmp, _ := strconv.ParseFloat(memlValue, 64)
		pod_Lmemory = tools.MiBToBytes(tmp)
	}

	if (pod_Memory > int64(avaMem)) || (pod.Cpu > ac_str) || (pod.Gpu > ag_str) {
		err := errors.New("sources required are larger than the avaliable!")
		c.JSON(http.StatusBadRequest, gin.H{
			"code: ":    http.StatusBadRequest,
			"message: ": err.Error(),
		})
		glog.Error("sources required are larger than the avaliable!")
		return
	}

	if (pod_Memory > pod_Lmemory) || (pod.Cpu > pod.Cpulim) || (pod.Gpu > pod.Gpulim) {
		err := errors.New("sources required are larger than the limited!")
		c.JSON(http.StatusBadRequest, gin.H{
			"code: ":    http.StatusBadRequest,
			"message: ": err.Error(),
		})
		glog.Error("sources required are larger than the limited!")
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
	// create pod
	pod_container, err := tools.CreatePod(pod, newPod)
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
