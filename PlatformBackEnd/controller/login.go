package controller

import (
	"PlatformBackEnd/data"
	"PlatformBackEnd/tools"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

func Login(c *gin.Context) {
	var user data.User
	err_bind := c.ShouldBindJSON(&user)
	if err_bind != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code: ":    http.StatusBadRequest,
			"message: ": fmt.Sprintf("Invalid request payload, err is %v", err_bind.Error()),
		})
		glog.Error("Method RegisterHandler gets invalid request payload")
		return
	}

	users, err := tools.LoadUsers("/home/gpu-server/set_k8s/biyesheji/PlatformBackEnd/test/users.json")
	if err != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code: ":  http.StatusMethodNotAllowed,
			"error: ": err.Error(),
		})
		glog.Error("Failed to load saved users info")
		return
	}

	var flag = 1

	for i := range users {
		if user.Username == users[i].Username && user.Password == users[i].Password {
			tools.GenerateToken(c, users[i])
			flag = 0
			break
		}
	}

	if flag == 1 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code: ":  http.StatusUnauthorized,
			"message": "Invalid credentials",
		})
	}
}
