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
	_ = flag.String("userport", ":8080", "the port to listen the platform")

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

	// Admin Opts
	router.POST("/login", controller.Login)
	router.GET("/getuser_notoken", controller.GetUserInfo_NoToken)
	router.POST("/modify_user", controller.ModifyUser)
	router.Static("/logs", flag.Lookup("log_dir").Value.String())

	api := router.Group("/api")
	api.Use(tools.JWTAuth())
	{
		// User Opts
		router.POST("/registe_user", controller.RegisterUser)
		router.POST("/delete_user", controller.DeleteUser)
		router.GET("/get_alluser", controller.GetAllUsers)

		// Dir Opts
		router.POST("/search_dir", controller.GetDirInfo)
		router.POST("/create_dir", controller.CreateDir)
		router.POST("/delete_dir", controller.DeleteDir)

		// Create Image
		router.POST("/image", controller.CreateImage)

		// Pod Opts
		router.POST("/create_pod", controller.CreatePod)
		router.POST("/delete_pod", controller.DeletePod)
		router.POST("/monite_k8s", controller.MonitorK8SResource)
		router.GET("/ws", controller.GetContainerData)

		// Data of model training Opts
		router.POST("/create_data", controller.GetModelLogData)
		router.POST("/delete_data", controller.DeleteModelLogData)

		// Handle Dir
		group := router.Group("/file")
		{
			group.POST("/list", controller.GetAllFiles)
			group.POST("/delete", controller.DeleteFile)
		}

		// Load file
		router.POST("/upload", controller.UploadFile)

	}

	router.Run(flag.Lookup("userport").Value.String())
}
