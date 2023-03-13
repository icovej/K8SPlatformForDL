package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"the_second_go/tools"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	gpuMetricName = "nvidia.com/gpu"
	srcfilepath = ""
)

type gpuUsage struct {
	ContainerName string    `json:"container_name"`
	Time          time.Time `json:"time"`
	GPUs          int64     `json:"gpus"`
}

// 初始化Docker客户端
func initDocker() (*client.Client, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		glog.Error("Failed to init docker client, the error is %s", err)
		return nil, err
	}
	
	_, err = dockerClient.Ping(context.Background())
	if err != nil {
		glog.Error("Failed to connect to docker client, the error is %s", err)
		return nil, err
	}

	glog.Info("Succeed to init docker client.")
	return dockerClient, nil
}

// 初始化Kubernetes客户端
func initK8S() (*kubernetes.clientSet, error){
	// 加载kubeconfog
	kubeConfig, errConfig := rest.InClusterConfig()
	if errConfig != nil {
		glog.Error("Failed to get kubeconfig, the error is %s", errConfig)
		return nil, errConfig
	}

	// 创建k8s客户端
	kubeClient, errClient := kubernetes.NewForConfig(kubeConfig)
	if errClient != nil {
		glog.Error("Failed to init kubeClient, the error is %s", errClient)
		return nil, errClient
	}

	glog.Info("Succeed to init kubeClient")
	return kubeClient, nil
}

// 制作镜像
func createImage(filepath string) {
	tools.ExecCommand("docker build -t ", filepath)
}

func main() {
	// 初始化Docker客户端
	dockerClient, err_docker := initDocker()
	defer dockerClient.Close()

	// 初始化Kubernetes客户端
	kubeClient, err_k8s := initK8S()

	// 初始化Gin框架
	router := gin.Default()

	// 镜像打包和拉取
	router.POST("/image", func(c *gin.Context) {
		// 用户指定的dockerfile保存路径
		dstFilepath := c.PostForm("dst_path")
		tools.CopyFile(srcfilepath, dstFilepath)

		// 选择基础镜像：Ubuntu/CentOS
		// os_version格式为: ubuntu:18.04
		osVersion := c.PostForm("os_version")
		cmd := "FROM " + osVersion + "\n"
		tools.writeAtHead(dstFilepath, cmd)
		
		// 选择的python版本
		pyVersion := c.PostFormArray("python_version")
		tools.writeAtTail(dstFilepath, pyVersion)

		// 选择镜像, 先拉取, 再将镜像名写入dockerfile中
		imageArray := c.PostFormArray("image_name_choose")
		imageNum := len(imageArray)
		glog.Info("The num of image choosen by user is %s", imageNum)
		for i := range imageArray {
			reader, err := dockerClient.ImagePull(context.Background(), imageArray[i], types.ImagePullOptions{})
			if err != nil {
				glog.Error("Failed to pull docker image %s, the error is %s", imageArray[i], err)
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			tools.WriteAtTail(dstFilepath, imageArray[i])
			glog.Info("Succeed to pull docker image %s", imageArray[i])
			defer reader.Close()
		}

		// 调用exec执行dockerfile，创建用户自定义镜像
		imageName := c.PostForm("image_name_user")
		cmd := "docker build -f "
		err_exec := tools.ExecCommand(cmd, dstFilepath, " -t ", imageName)
		if err_exec != nil {
			glog.Error("Failed to exec docker build, the error is %s", err_exec)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Succeed to build image: %s", imageName),
		})
	})

	// 容器创建
	router.POST("/containers", func(c *gin.Context) {
		containerName := c.PostForm("container_name")
		memoryStr := c.PostForm("memory")
		cpuStr := c.PostForm("cpu")
		gpuStr := c.PostForm("gpu")
		imageName := c.PostForm("image")

		// 解析内存、CPU和GPU数
		memReq, err_mem := resource.ParseQuantity(memoryStr)
		if err_mem != nil {
			glog.Error("Failed to parse memReq, the error is %s", err_mem)
			c.AbortWithError(http.StatusBadRequest, err_mem)
			return
		}
		cpuReq, err_cpu := resource.ParseQuantity(cpuStr)
		if err_cpu != nil {
			glog.Error("Failed to parse cpu, the error is %s", err_cpu)
			c.AbortWithError(http.StatusBadRequest, err_cpu)
			return
		}
		gpuReq, err_gpu := resource.ParseQuantity(gpuStr)
		if err_gpu != nil {
			glog.Error("Failed to parse gpu, the error is %s", err_gpu)
			c.AbortWithError(http.StatusBadRequest, err_gpu)
			return
		}

		// 获取当前可用的Mem、CPU和PU
		var avaGPU int64
		m := make(map[int]uint64)
		avaMem, avaCPU, m, _ := tools.GetAvailableMemoryAndGPU()
		for i := range m {
			avaGPU += m[i]
		}

		// 比较用户请求数和可用数
		if memReq > avaMem | cpuReq > avaCPU | gpuReq > avaGPU {
			glog.Error("Failed to alloc sources to create pod, because the free sources are limited!")
			return
		}


		// 创建容器
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
			Name: containerName,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  containerName,
					Image: containerName,
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: memory,
							v1.ResourceName(gpuMetricName): gpu,
						},
					},
				},
			},
		},
	}

	// 创建Pod
	_, err = kubeClient.CoreV1().Pods("default").Create(context.Background(), pod, metav1.CreateOptions{})
	if err != nil {
		log.Println(err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Container %s has been created.", containerName),
	})

	// GPU使用情况查询
	router.GET("/gpu_usage", func(c *gin.Context) {
		var gpuUsages []gpuUsage

		// 获取所有Pod
		podList, err := kubeClient.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})
		if err != nil {
			log.Println(err)
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// 遍历所有Pod
		for _, pod := range podList.Items {
			// 遍历Pod中的所有容器
			for _, container := range pod.Spec.Containers {
				// 获取容器GPU使用情况
				metrics, err := kubeClient.MetricsV1beta1().PodMetricses("default").Get(context.Background(), pod.Name, metav1.GetOptions{})
				if err != nil {
					log.Println(err)
					continue
				}
				for _, containerMetrics := range metrics.Containers {
					if containerMetrics.Name == container.Name {
						gpuUsage := gpuUsage{
							ContainerName: container.Name,
							Time:          time.Now(),
							GPUs:          containerMetrics.Usage[v1.ResourceName(gpuMetricName)].Value(),
						}
						gpuUsages = append(gpuUsages, gpuUsage)
					}
				}
			}
		}

		// 返回所有GPU使用情况
		c.JSON(http.StatusOK, gin.H{
			"gpu_usage": gpuUsages,
		})
	})

	// GPU使用情况查询（历史记录）
	router.GET("/gpu_usage/:container_name", func(c *gin.Context) {
		containerName := c.Param("container_name")

		// 获取所有Pod
		podList, err := kubeClient.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})
		if err != nil {
			log.Println(err)
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// 遍历所有Pod
		var gpuUsages []gpuUsage
		for _, pod := range podList.Items {
			// 遍历Pod中的所有容器
			for _, container := range pod.Spec.Containers {
				// 如果容器名匹配，则获取GPU使用情况
				if container.Name == containerName {
					metricsList, err := kubeClient.MetricsV1beta1().PodMetricses("default").List(context.Background(), metav1.ListOptions{})
					if err != nil {
						log.Println(err)
						continue
					}
					for _, metrics := metricsList.Items
						sort.Slice(metrics, func(i, j int) bool {
							return metrics[i].Timestamp.Time.Before(metrics[j].Timestamp.Time)
						})

						var gpuUsagesBefore []gpuUsage
						var gpuUsagesAfter []gpuUsage

						for _, containerMetrics := range metrics {
							if containerMetrics.Name == containerName {
								gpuUsage := gpuUsage{
									ContainerName: container.Name,
									Time:          containerMetrics.Timestamp.Time,
									GPUs:          containerMetrics.Containers[0].Usage[v1.ResourceName(gpuMetricName)].Value(),
								}
								if gpuUsage.Time.Before(currentTime.Add(-30 * time.Second)) {
									gpuUsagesBefore = append(gpuUsagesBefore, gpuUsage)
								} else {
									gpuUsagesAfter = append(gpuUsagesAfter, gpuUsage)
								}
							}
						}

						// 返回指定容器的历史GPU使用情况
						c.JSON(http.StatusOK, gin.H{
							"gpu_usage_before": gpuUsagesBefore,
							"gpu_usage_after":  gpuUsagesAfter,
						})
						return
					}
				}
			}
		}

		c.JSON(http.StatusNotFound, gin.H{
			"message": fmt.Sprintf("Container %s not found.", containerName),
		})
	})

	router.Run(":8080")
}