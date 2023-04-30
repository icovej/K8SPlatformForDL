package controller

import (
	"PlatformBackEnd/data"
	"PlatformBackEnd/tools"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

func Login(c *gin.Context) {
	var user data.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(data.API_PARAMETER_ERROR, gin.H{
			"code: ":    data.API_PARAMETER_ERROR,
			"message: ": fmt.Sprintf("Invalid request payload, err is %v", err.Error()),
		})
		glog.Error("Method RegisterHandler gets invalid request payload")
		return
	}

	users, err := tools.LoadUsers(data.UserFile)
	if err != nil {
		c.JSON(data.OPERATION_FAILURE, gin.H{
			"code: ":  data.OPERATION_FAILURE,
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
		c.JSON(data.OPERATION_FAILURE, gin.H{
			"code: ":  data.OPERATION_FAILURE,
			"message": "Invalid credentials",
		})
	}
}
