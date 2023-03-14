package main

import (
	"the_second_go/controller"
	"the_second_go/tools"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化Kubernetes客户端
	kubeClient, err_k8s := tools.InitK8S()

	// 初始化Gin框架
	router := gin.Default()

	// 获取api操作信息
	router.GET("/operation", controller.OperationInfo)

	// 账号注册
	router.POST("/register", controller.RegisterHandler)

	// 账号登陆
	router.POST("/login", controller.Login)

	// 查询目录容量
	router.GET("/search_dir", controller.GetDirInfo)

	// 创建镜像
	router.POST("/image", controller.CreateImage)

	// 创建Pod
	router.POST("/pod", controller.CreatePod)

}
