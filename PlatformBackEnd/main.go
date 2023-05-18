package main

import (
	"PlatformBackEnd/controller"
	"PlatformBackEnd/data"
	"PlatformBackEnd/tools"
	"flag"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/polling"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
)

func main() {
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
	router.Use(tools.Core())

	// Init Websocket
	var socketconfig = &engineio.Options{
		PingTimeout:  7 * time.Second,
		PingInterval: 5 * time.Second,
		Transports: []transport.Transport{
			&polling.Transport{
				Client: &http.Client{
					Timeout: time.Minute,
				},
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			},
			&websocket.Transport{
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			},
		},
	}
	server := socketio.NewServer(socketconfig)
	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		glog.Infof("Client connected: %v", s.ID())

		return nil
	})
	server.OnEvent("/", "data", func(s socketio.Conn, data map[string]interface{}) {
		value, _ := data["value"].(string)
		glog.Infof("value = %v", value)
		go func(s socketio.Conn) {
			for {
				tools.GetContainerData(s, value)
				time.Sleep(time.Second)
			}
		}(s)
	})
	go func() {
		if err := server.Serve(); err != nil {
			glog.Infof("socketio listen error: %s\n", err)
		}
	}()
	defer server.Close()
	router.Use(tools.Core())
	router.GET("/socket.io/*any", gin.WrapH(server))
	router.POST("/socket.io/*any", gin.WrapH(server))

	// set the max memory of file uploaded
	//router.MaxMultipartMemory = 8 << 30 // 8GB

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
		router.POST("/get_pod", controller.GetK8SPod)
		router.GET("/get_namespace", controller.GetK8SNamespace)
		router.POST("/gpu_share", controller.GetGPUShareData)
		router.GET("/gpu_node", controller.GetK8SNodeGPU)

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
