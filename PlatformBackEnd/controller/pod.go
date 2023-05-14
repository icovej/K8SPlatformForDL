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
	err := c.ShouldBindJSON(&pod)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    data.API_PARAMETER_ERROR,
			"message": fmt.Sprintf("Invalid request payload, err is %v", err.Error()),
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
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": err.Error(),
		})
		glog.Error("sources required are larger than the avaliable!")
		return
	}

	if (pod_Memory > pod_Lmemory) || (pod.Cpu > pod.Cpulim) || (pod.Gpu > pod.Gpulim) {
		err := errors.New("sources required are larger than the limited!")
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": err.Error(),
		})
		glog.Error("sources required are larger than the limited!")
		return
	}

	// Parse mem、CPU and GPU to k8s mod
	memReq, err := resource.ParseQuantity(pod.Memory)
	if err != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": err.Error(),
		})
		glog.Errorf("Failed to parse mem, the error is %v", err)
		return
	}
	memLim, err := resource.ParseQuantity(pod.Memlim)
	if err != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": err.Error(),
		})
		glog.Errorf("Failed to parse mem, the error is %v", err)
		return
	}

	cpuReq, err := resource.ParseQuantity(pod.Cpu)
	if err != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": err.Error(),
		})
		glog.Errorf("Failed to parse cpu, the error is %v", err)
		return
	}
	cpuLim, err := resource.ParseQuantity(pod.Cpulim)
	if err != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": err.Error(),
		})
		glog.Errorf("Failed to parse mem, the error is %v", err)
		return
	}

	gpuReq, err := resource.ParseQuantity(pod.Gpu)
	if err != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": err.Error(),
		})
		glog.Errorf("Failed to parse gpu, the error is %v", err)
		return
	}
	gpuLim, err := resource.ParseQuantity(pod.Gpulim)
	if err != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": err.Error(),
		})
		glog.Errorf("Failed to parse mem, the error is %v", err)
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
			"code":    data.OPERATION_FAILURE,
			"message": err.Error(),
		})
		glog.Errorf("Failed to create pod %v", err)
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{
		"code":    data.SUCCESS,
		"message": fmt.Sprintf("Succeed to create pod, its name is %v", pod_container.GetObjectMeta().GetName()),
	})
}
