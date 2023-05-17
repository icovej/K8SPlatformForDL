package controller

import (
	"PlatformBackEnd/tools"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

// var upgrader = websocket.Upgrader{
// 	ReadBufferSize:  1024,
// 	WriteBufferSize: 1024,
// }

func GetContainerData(c *gin.Context) {
	upGrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		glog.Errorf("Failed to upgrade WebSocket: %v", err.Error())
		return
	}

	glog.Info("Succeed to build websocket, %v", conn.RemoteAddr())

	go func(conn *websocket.Conn) {
		for {
			tools.GetContainerData(conn)
			time.Sleep(time.Second)
		}
	}(conn)
	// log.Println("连接建立成功", conn.RemoteAddr())
	// for {
	// 	tools.GetContainerData(conn)
	// 	time.Sleep(time.Second)
	// }
}
