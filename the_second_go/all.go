package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	gpuMetricName = "nvidia.com/gpu"
)

type gpuUsage struct {
	ContainerName string    `json:"container_name"`
	Time          time.Time `json:"time"`
	GPUs          int64     `json:"gpus"`
}

func main() {
	// 初始化Docker客户端
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal(err)
	}
	defer dockerClient.Close()

	// 初始化Kubernetes客户端
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		log.Fatal(err)
	}

	// 初始化Gin框架
	router := gin.Default()

	// 镜像打包和拉取
	router.POST("/images", func(c *gin.Context) {
		imageName := c.PostForm("image_name")
		reader, err := dockerClient.ImagePull(context.Background(), imageName, types.ImagePullOptions{})
		if err != nil {
			log.Println(err)
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		defer reader.Close()
		if _, err := client.CopyFromContainer(context.Background(), imageName, types.CopyToContainerOptions{}, "/"); err != nil {
			log.Println(err)
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Image %s has been pulled and packed.", imageName),
		})
	})

	// 容器创建
	router.POST("/containers", func(c *gin.Context) {
		containerName := c.PostForm("container_name")
		memoryStr := c.PostForm("memory")
		gpuStr := c.PostForm("gpu")

		// 解析内存和GPU数
		memory, err := resource.ParseQuantity(memoryStr)
		if err != nil {
			log.Println(err)
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		gpu, err := resource.ParseQuantity(gpuStr)
		if err != nil {
			log.Println(err)
			c.AbortWithError(http.StatusBadRequest, err)
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