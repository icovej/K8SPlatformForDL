package main

import (
	"PlatformBackEnd/controller"
	"PlatformBackEnd/data"
	"PlatformBackEnd/tools"
	"flag"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

func main() {
	// parse cmdline
	var srcfilepath = flag.String("srcfilepath", "", "the original dockerfile path")
	data.Srcfilepath = *srcfilepath

	var logdir = flag.String("logdir", "", "The path to save glog")
	flag.Lookup("log_dir").Value.Set(*logdir)

	var port = flag.String("userport", ":8080", "the port to listen the platform")
	flag.Parse()
	fmt.Println(*port)

	defer glog.Flush()
	glog.Info("Succeed to start platform")

	_ = tools.CreateFile(data.UserFile)

	// Init Gin
	router := gin.Default()

	router.Use(tools.Core())

	// Get API information
	router.GET("/operation", controller.OperationInfo)

	// Registe
	router.POST("/register", controller.RegisterHandler)

	// Login
	router.POST("/login", controller.Login)

	// Query Dir Info
	router.GET("/search_dir", controller.GetDirInfo)

	// Create Image
	router.POST("/image", controller.CreateImage)

	// Create Pod
	router.POST("/pod", controller.CreatePod)

	// Get data of model training
	router.POST("/data", controller.GetData)

	// Monite Pod
	router.POST("/monitor", controller.MonitorPod)

	// Handle Dir
	router.GET("/list", controller.GetAllFiles)
	router.DELETE("/delete", controller.DeleteFile)

	// Get container data
	router.POST("/ws", controller.GetContainerData)

	// Load file
	router.GET("/upload", controller.UploadFile)

	router.Run(*port)
	// 192.168.10.11
	// router.Run("192.168.10.11:8080")
}
