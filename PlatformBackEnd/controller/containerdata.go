package controller

import (
	"PlatformBackEnd/tools"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func GetContainerData(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		glog.Errorf("Error upgrading to WebSocket: %v", err.Error())
		return
	}

	go func(conn *websocket.Conn) {
		for {
			tools.GetContainerData(conn)
			time.Sleep(time.Second)
		}
	}(conn)
}
