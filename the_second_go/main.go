package main

import (
	"flag"
	"the_second_go/controller"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

func main() {
	// 启动glog
	flag.Parse()
	defer glog.Flush()

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

	// 获取训练模型的损失值和正确率
	router.POST("/data", controller.GetData)
}
