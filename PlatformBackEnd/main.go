package main

import (
	"PlatformBackEnd/controller"
	"PlatformBackEnd/data"
	"PlatformBackEnd/tools"
	"flag"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

func main() {
	// parse cmdline
	// var srcfilepath = flag.String("srcfilepath", "", "the original dockerfile path")
	// data.Srcfilepath = *srcfilepath
	// fmt.Printf("x = %v", *srcfilepath)
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		glog.Errorf("failed to load location: %v", err.Error())
	}
	time.Local = loc

	var logdir = flag.String("logdir", "", "The path to save glog")
	flag.Lookup("log_dir").Value.Set(*logdir)
	var port = flag.String("userport", ":8080", "the port to listen the platform")

	flag.Parse()
	defer glog.Flush()
	glog.Info("Succeed to start platform")

	_ = tools.CreateFile(data.UserFile)

	// Init Gin
	router := gin.Default()

	// set the max memory of file uploaded
	//router.MaxMultipartMemory = 8 << 30 // 8GB

	router.Use(tools.Core())

	// Get API information
	router.GET("/operation", controller.OperationInfo)

	// Login
	router.POST("/login", controller.Login)

	api := router.Group("/api")
	api.Use(tools.JWTAuth())
	{
		// Registe
		router.POST("/register", controller.RegisterHandler)

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
		group := router.Group("/file")
		{
			group.GET("/list", controller.GetAllFiles)
			group.DELETE("/delete", controller.DeleteFile)
		}

		// Get container data
		router.POST("/ws", controller.GetContainerData)

		// Load file
		router.POST("/upload", controller.UploadFile)

		// Get all user
		router.GET("/getuser", controller.GetAllUsers)
	}

	router.Run(*port)
}
